package prototype

import (
	"encoding/base64"
	"fmt"
	"github.com/go-leo/gox/stringx"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func Unmarshal(data []byte, tgt any, opts *options) error {
	var d decodeState
	d.data = data
	d.off = 0
	return unmarshal(&d, tgt, opts)
}

func unmarshal(d *decodeState, tgt any, opts *options) error {
	tgtVal := reflect.ValueOf(tgt)
	if tgtVal.Kind() != reflect.Pointer || tgtVal.IsNil() {
		return &InvalidTargetError{Type: reflect.TypeOf(tgt)}
	}

	d.scan.reset()
	d.scanWhile(scanSkipSpace)
	// We decode tgtVal not tgtVal.Elem because the Unmarshaler interface
	// test must be applied at the top level of the value.
	return value(d, tgtVal, opts)
}

// value consumes a JSON value from d.data[d.off-1:], decoding into v, and
// reads the following byte ahead. If v is invalid, the value is discarded.
// The first byte of the value has been read already.
func value(d *decodeState, v reflect.Value, opts *options) error {
	switch d.opcode {
	default:
		return ErrPhase
	case scanBeginArray:
		if v.IsValid() {
			if err := array(d, v, opts); err != nil {
				return err
			}
		} else {
			d.skip()
		}
		d.scanNext()
	case scanBeginObject:
		if v.IsValid() {
			if err := object(d, v, opts); err != nil {
				return err
			}
		} else {
			d.skip()
		}
		d.scanNext()
	case scanBeginLiteral:
		// All bytes inside literal return scanContinue op code.
		start := d.readIndex()
		d.rescanLiteral()

		if v.IsValid() {
			item := d.data[start:d.readIndex()]
			if err := literalStore(item, v, false); err != nil {
				return err
			}
		}
	}
	return nil
}

// array consumes an array from d.data[d.off-1:], decoding into v.
// The first byte of the array ('[') has been read already.
func array(d *decodeState, v reflect.Value, opts *options) error {
	// Check for unmarshaler.
	v = indirect(v, false)
	// Check type of target.
	switch v.Kind() {
	case reflect.Interface:
		return setArrayInterface(d, v)
	case reflect.Array, reflect.Slice:
		return setArrayOrSlice(d, v, opts)
	default:
		return &CloneError{Value: "array", Type: v.Type()}
	}
}

func setArrayOrSlice(d *decodeState, tv reflect.Value, opts *options) error {
	i := 0
	for {
		// Look ahead for ] - can only happen on first iteration.
		d.scanWhile(scanSkipSpace)
		if d.opcode == scanEndArray {
			break
		}

		// Get element of array, growing if necessary.
		if tv.Kind() == reflect.Slice {
			// Grow slice if necessary
			if i >= tv.Cap() {
				newcap := tv.Cap() + tv.Cap()/2
				if newcap < 4 {
					newcap = 4
				}
				newv := reflect.MakeSlice(tv.Type(), tv.Len(), newcap)
				reflect.Copy(newv, tv)
				tv.Set(newv)
			}
			if i >= tv.Len() {
				tv.SetLen(i + 1)
			}
		}

		if i < tv.Len() {
			// Decode into element.
			if err := value(d, tv.Index(i), opts); err != nil {
				return err
			}
		} else {
			// Ran out of fixed array: skip.
			if err := value(d, reflect.Value{}, opts); err != nil {
				return err
			}
		}
		i++

		// Next token must be , or ].
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		if d.opcode == scanEndArray {
			break
		}
		if d.opcode != scanArrayValue {
			return ErrPhase
		}
	}

	if i < tv.Len() {
		if tv.Kind() == reflect.Array {
			// Array. Zero the rest.
			z := reflect.Zero(tv.Type().Elem())
			for ; i < tv.Len(); i++ {
				tv.Index(i).Set(z)
			}
		} else {
			tv.SetLen(i)
		}
	}
	if i == 0 && tv.Kind() == reflect.Slice {
		tv.Set(reflect.MakeSlice(tv.Type(), 0, 0))
	}
	return nil
}

func setArrayInterface(d *decodeState, v reflect.Value) error {
	if v.NumMethod() == 0 {
		// Decoding into nil interface? Switch to non-reflect code.
		ai, err := arrayInterface(d)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(ai))
		return nil
	}
	// Otherwise it's invalid.
	return &CloneError{Value: "array", Type: v.Type()}
}

// object consumes an object from d.data[d.off-1:], decoding into v.
// The first byte ('{') of the object has been read already.
func object(d *decodeState, tgtVal reflect.Value, opts *options) error {
	tgtVal = indirect(tgtVal, false)
	t := tgtVal.Type()

	// Decoding into nil interface? Switch to non-reflect code.
	if tgtVal.Kind() == reflect.Interface && tgtVal.NumMethod() == 0 {
		oi, err := objectInterface(d)
		if err != nil {
			return err
		}
		tgtVal.Set(reflect.ValueOf(oi))
		return nil
	}

	var fields structFields

	// Check type of target:
	//   struct or
	//   map[T1]T2 where T1 is string, an integer type,
	//             or an encoding.TextUnmarshaler
	switch tgtVal.Kind() {
	case reflect.Map:
		// Map key must either have string kind, have an integer kind,
		// or be an encoding.TextUnmarshaler.
		switch t.Key().Kind() {
		case reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		default:
			if !reflect.PointerTo(t.Key()).Implements(textUnmarshalerType) {
				return &CloneError{Value: "object", Type: t}
			}
		}
		if tgtVal.IsNil() {
			tgtVal.Set(reflect.MakeMap(t))
		}
	case reflect.Struct:
		fields = cachedTypeFields(t, opts, opts.TargetTagKey)
		// ok
	default:
		return &CloneError{Value: "object", Type: t}
	}

	var mapElem reflect.Value

	for {
		// Read opening " of string key or closing }.
		d.scanWhile(scanSkipSpace)
		if d.opcode == scanEndObject {
			// closing } - can only happen on first iteration.
			break
		}
		if d.opcode != scanBeginLiteral {
			return ErrPhase
		}

		// Read key.
		start := d.readIndex()
		d.rescanLiteral()
		item := d.data[start:d.readIndex()]
		key, ok := unquoteBytes(item)
		if !ok {
			return ErrPhase
		}

		var subv reflect.Value

		if tgtVal.Kind() == reflect.Map {
			elemType := t.Elem()
			if !mapElem.IsValid() {
				mapElem = reflect.New(elemType).Elem()
			} else {
				mapElem.Set(reflect.Zero(elemType))
			}
			subv = mapElem
		} else {
			var tgtField *field
			if i, ok := fields.dominantsNameIndex[string(key)]; ok {
				// Found an exact Nil match.
				tgtField = &fields.dominants[i]
			} else {
				// Fall back to the expensive case-insensitive
				// linear search.
				for i := range fields.dominants {
					ff := &fields.dominants[i]
					if ff.equalFold(ff.nameBytes, key) {
						tgtField = ff
						break
					}
				}
			}
			if tgtField != nil {
				subv = tgtVal
				for _, i := range tgtField.index {
					if subv.Kind() == reflect.Pointer {
						if subv.IsNil() {
							// If a struct embeds a pointer to an unexported type,
							// it is not possible to set a newly allocated value
							// since the field is unexported.
							//
							// See https://golang.org/issue/21357
							if !subv.CanSet() {
								d.saveError(fmt.Errorf("json: cannot set embedded pointer to unexported struct: %v", subv.Type().Elem()))
								// Invalidate subv to ensure d.value(subv) skips over
								// the JSON value without assigning it to subv.
								subv = reflect.Value{}
								break
							}
							subv.Set(reflect.New(subv.Type().Elem()))
						}
						subv = subv.Elem()
					}
					subv = subv.Field(i)
				}
			} else if d.disallowUnknownFields {
				d.saveError(fmt.Errorf("json: unknown field %q", key))
			}
		}

		// Read : before value.
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		if d.opcode != scanObjectKey {
			return ErrPhase
		}
		d.scanWhile(scanSkipSpace)

		if err := value(d, subv, opts); err != nil {
			return err
		}

		// Write value back to map;
		// if using struct, subv points into struct already.
		if tgtVal.Kind() == reflect.Map {
			kt := t.Key()
			var kv reflect.Value
			switch {
			case reflect.PointerTo(kt).Implements(textUnmarshalerType):
				kv = reflect.New(kt)
				if err := literalStore(item, kv, true); err != nil {
					return err
				}
				kv = kv.Elem()
			case kt.Kind() == reflect.String:
				kv = reflect.ValueOf(key).Convert(kt)
			default:
				switch kt.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					s := string(key)
					n, err := strconv.ParseInt(s, 10, 64)
					if err != nil || reflect.Zero(kt).OverflowInt(n) {
						d.saveError(&CloneError{Value: "number " + s, Type: kt})
						break
					}
					kv = reflect.ValueOf(n).Convert(kt)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					s := string(key)
					n, err := strconv.ParseUint(s, 10, 64)
					if err != nil || reflect.Zero(kt).OverflowUint(n) {
						d.saveError(&CloneError{Value: "number " + s, Type: kt})
						break
					}
					kv = reflect.ValueOf(n).Convert(kt)
				default:
					panic("json: Unexpected key type") // should never occur
				}
			}
			if kv.IsValid() {
				tgtVal.SetMapIndex(kv, subv)
			}
		}

		// Next token must be , or }.
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		//if d.errorContext != nil {
		//	// Reset errorContext to its original state.
		//	// Keep the same underlying array for FieldStack, to reuse the
		//	// space and avoid unnecessary allocs.
		//	d.errorContext.FieldStack = d.errorContext.FieldStack[:len(origErrorContext.FieldStack)]
		//	d.errorContext.Struct = origErrorContext.Struct
		//}
		if d.opcode == scanEndObject {
			break
		}
		if d.opcode != scanObjectValue {
			return ErrPhase
		}
	}
	return nil
}

// convertNumber converts the number literal s to a float64 or a Number
// depending on the setting of d.useNumber.
func convertNumber(s string) (any, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, &CloneError{Value: "number " + s, Type: reflect.TypeOf(0.0)}
	}
	return f, nil
}

// literalStore decodes a literal stored in item into v.
//
// fromQuoted indicates whether this literal came from unwrapping a
// string from the ",string" struct tag option. this is used only to
// produce more helpful error messages.
func literalStore(item []byte, v reflect.Value, fromQuoted bool) error {
	if len(item) == 0 {
		return fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal %q into %v", item, v.Type())
	}
	isNull := item[0] == 'n' // null
	pv := indirect(v, isNull)
	return setSampleValue(item, pv, fromQuoted)
}

func setSampleValue(item []byte, v reflect.Value, fromQuoted bool) error {
	switch c := item[0]; c {
	case 'n':
		// null
		// The main parser checks that only true and false can reach here,
		// but if this was a quoted string input, it could be anything.
		if fromQuoted && string(item) != "null" {
			return fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal %q into %v", item, v.Type())
		}
		return setNil(v)
	case 't', 'f':
		// true, false
		value := item[0] == 't'
		// The main parser checks that only true and false can reach here,
		// but if this was a quoted string input, it could be anything.
		if fromQuoted && string(item) != "true" && string(item) != "false" {
			return fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal %q into %v", item, v.Type())
		}
		return setBool(item, v, fromQuoted, value)
	case '"':
		// string
		s, ok := unquoteBytes(item)
		if !ok {
			if fromQuoted {
				return fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal %q into %v", item, v.Type())
			}
			return ErrPhase
		}
		return setString(item, v, s)

	default:
		// number
		if c != '-' && (c < '0' || c > '9') {
			if fromQuoted {
				return fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal %q into %v", item, v.Type())
			}
			return ErrPhase
		}
		return setNumber(item, v)
	}
}

func setNumber(item []byte, v reflect.Value) error {
	s := string(item)
	switch v.Kind() {
	default:
		if setTypeNumber(v, s) {
			return nil
		}
		return &CloneError{Value: "number", Type: v.Type()}
	case reflect.Interface:
		return setInterfaceNumber(v, s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return &CloneError{Value: "number " + s, Type: v.Type()}
		}
		return setIntNumber(v, n, s)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return &CloneError{Value: "number " + s, Type: v.Type()}
		}
		return setUintNumber(v, n, s)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return &CloneError{Value: "number " + s, Type: v.Type()}
		}
		return setFloatNumber(v, n, s)
	}
}

func setString(item []byte, v reflect.Value, s []byte) error {
	switch v.Kind() {
	default:
		return &CloneError{Value: "string", Type: v.Type()}
	case reflect.Slice:
		if v.Type().Elem().Kind() != reflect.Uint8 {
			return &CloneError{Value: "string", Type: v.Type()}
		}
		b := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
		n, err := base64.StdEncoding.Decode(b, s)
		if err != nil {
			return err
		}
		v.SetBytes(b[:n])
	case reflect.String:
		if v.Type() == numberType && !stringx.IsValidNumber(string(s)) {
			return fmt.Errorf("json: invalid number literal, trying to unmarshal %q into Number", item)
		}
		v.SetString(string(s))
	case reflect.Interface:
		if v.NumMethod() == 0 {
			v.Set(reflect.ValueOf(string(s)))
		} else {
			return &CloneError{Value: "string", Type: v.Type()}
		}
	}
	return nil
}

func setBool(item []byte, v reflect.Value, fromQuoted bool, value bool) error {
	switch v.Kind() {
	default:
		if fromQuoted {
			return fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal %q into %v", item, v.Type())
		} else {
			return &CloneError{Value: "bool", Type: v.Type()}
		}
	case reflect.Bool:
		v.SetBool(value)
	case reflect.Interface:
		if v.NumMethod() == 0 {
			v.Set(reflect.ValueOf(value))
		} else {
			return &CloneError{Value: "bool", Type: v.Type()}
		}
	}
	return nil
}

func setNil(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Map, reflect.Slice:
		v.Set(reflect.Zero(v.Type()))
		// otherwise, ignore nil for primitives/string
	}
	return nil
}

func setTypeNumber(v reflect.Value, s string) bool {
	if v.Kind() == reflect.String && v.Type() == numberType {
		// s must be a valid number, because it's
		// already been tokenized.
		v.SetString(s)
		return true
	}
	return false
}

func setInterfaceNumber(v reflect.Value, s string) error {
	n, err := convertNumber(s)
	if err != nil {
		return err
	}
	if v.NumMethod() != 0 {
		return &CloneError{Value: "number", Type: v.Type()}
	}
	v.Set(reflect.ValueOf(n))
	return nil
}

func setIntNumber(v reflect.Value, n int64, s string) error {
	if v.OverflowInt(n) {
		return &CloneError{Value: "number " + s, Type: v.Type()}
	}
	v.SetInt(n)
	return nil
}

func setUintNumber(v reflect.Value, n uint64, s string) error {
	if v.OverflowUint(n) {
		return &CloneError{Value: "number " + s, Type: v.Type()}
	}
	v.SetUint(n)
	return nil
}

func setFloatNumber(v reflect.Value, n float64, s string) error {
	if v.OverflowFloat(n) {
		return &CloneError{Value: "number " + s, Type: v.Type()}
	}
	v.SetFloat(n)
	return nil
}

// The xxxInterface routines build up a value to be stored
// in an empty interface. They are not strictly necessary,
// but they avoid the weight of reflection in this common case.

// valueInterface is like value but returns interface{}
func valueInterface(d *decodeState) (any, error) {
	switch d.opcode {
	default:
		return nil, ErrPhase
	case scanBeginArray:
		val, err := arrayInterface(d)
		d.scanNext()
		return val, err
	case scanBeginObject:
		val, err := objectInterface(d)
		d.scanNext()
		return val, err
	case scanBeginLiteral:
		val, err := literalInterface(d)
		return val, err
	}
}

// arrayInterface is like array but returns []interface{}.
func arrayInterface(d *decodeState) ([]any, error) {
	var v = make([]any, 0)
	for {
		// Look ahead for ] - can only happen on first iteration.
		d.scanWhile(scanSkipSpace)
		if d.opcode == scanEndArray {
			break
		}

		vi, err := valueInterface(d)
		if err != nil {
			return nil, err
		}
		v = append(v, vi)

		// Next token must be , or ].
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		if d.opcode == scanEndArray {
			break
		}
		if d.opcode != scanArrayValue {
			return nil, ErrPhase
		}
	}
	return v, nil
}

// objectInterface is like object but returns map[string]interface{}.
func objectInterface(d *decodeState) (map[string]any, error) {
	m := make(map[string]any)
	for {
		// Read opening " of string key or closing }.
		d.scanWhile(scanSkipSpace)
		if d.opcode == scanEndObject {
			// closing } - can only happen on first iteration.
			break
		}
		if d.opcode != scanBeginLiteral {
			return nil, ErrPhase
		}

		// Read string key.
		start := d.readIndex()
		d.rescanLiteral()
		item := d.data[start:d.readIndex()]
		key, ok := unquote(item)
		if !ok {
			return nil, ErrPhase
		}

		// Read : before value.
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		if d.opcode != scanObjectKey {
			return nil, ErrPhase
		}
		d.scanWhile(scanSkipSpace)

		// Read value.
		vi, err := valueInterface(d)
		if err != nil {
			return nil, ErrPhase
		}
		m[key] = vi
		// Next token must be , or }.
		if d.opcode == scanSkipSpace {
			d.scanWhile(scanSkipSpace)
		}
		if d.opcode == scanEndObject {
			break
		}
		if d.opcode != scanObjectValue {
			return nil, ErrPhase
		}
	}
	return m, nil
}

// literalInterface consumes and returns a literal from d.data[d.off-1:] and
// it reads the following byte ahead. The first byte of the literal has been
// read already (that's how the caller knows it's a literal).
func literalInterface(d *decodeState) (any, error) {
	// All bytes inside literal return scanContinue op code.
	start := d.readIndex()
	d.rescanLiteral()

	item := d.data[start:d.readIndex()]

	switch c := item[0]; c {
	case 'n': // null
		return nil, nil

	case 't', 'f': // true, false
		return c == 't', nil

	case '"': // string
		s, ok := unquote(item)
		if !ok {
			return nil, ErrPhase
		}
		return s, nil

	default: // number
		if c != '-' && (c < '0' || c > '9') {
			return nil, ErrPhase
		}
		n, err := convertNumber(string(item))
		if err != nil {
			d.saveError(err)
		}
		return n, nil
	}
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	var r rune
	for _, c := range s[2:6] {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return -1
		}
		r = r*16 + rune(c)
	}
	return r
}

// unquote converts a quoted JSON string literal s into an actual string t.
// The rules are different than for Go, so cannot use strconv.Unquote.
func unquote(s []byte) (t string, ok bool) {
	s, ok = unquoteBytes(s)
	t = string(s)
	return
}

func unquoteBytes(s []byte) (t []byte, ok bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return
	}
	s = s[1 : len(s)-1]

	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room? Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// If isNil is true, indirect stops at the first settable pointer so it
// can be set to nil.
func indirect(v reflect.Value, isNil bool) reflect.Value {
	// Issue #24153 indicates that it is generally not a guaranteed property
	// that you may round-trip a reflect.Value by calling Value.Addr().Elem()
	// and expect the value to still be settable for values derived from
	// unexported embedded struct fields.
	//
	// The logic below effectively does this when it first addresses the value
	// (to satisfy possible pointer methods) and continues to dereference
	// subsequent pointers as necessary.
	//
	// After the first round-trip, we set v back to the original value to
	// preserve the original RW flags contained in reflect.Value.
	v0 := v
	haveAddr := false

	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Pointer && v.Type().Name() != "" && v.CanAddr() {
		haveAddr = true
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Pointer && !e.IsNil() && (!isNil || e.Elem().Kind() == reflect.Pointer) {
				haveAddr = false
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Pointer {
			break
		}

		if isNil && v.CanSet() {
			break
		}

		// Prevent infinite loop if v is an interface pointing to its own address:
		//     var v interface{}
		//     v = &v
		if v.Elem().Kind() == reflect.Interface && v.Elem().Elem() == v {
			v = v.Elem()
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		if haveAddr {
			v = v0 // restore original value after round-trip Value.Addr().Elem()
			haveAddr = false
		} else {
			v = v.Elem()
		}
	}
	return v
}
