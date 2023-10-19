package prototype

import (
	"encoding"
	"fmt"
	"github.com/go-leo/gox/convx"
	"github.com/go-leo/gox/reflectx"
	"golang.org/x/exp/slices"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

func invalidValueEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	writeNull(e, fks, v, opts)
	return nil
}

func writeNull(e *encodeState, fks []string, v reflect.Value, _ encOpts) {
	printValue(fks, "null")
}

func printValue(fks []string, val any) {
	fmt.Println(strings.Join(fks, "."), "=", val)
}

func marshalerEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	m, ok := v.Interface().(Marshaler)
	if !ok {
		writeNull(e, fks, v, opts)
		return nil
	}
	b, err := m.MarshalJSON()
	if err != nil {
		return &MarshalerError{v.Type(), err, "MarshalJSON"}
	}
	printValue(fks, string(b))
	return nil
}

func addrMarshalerEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	va := v.Addr()
	if va.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	m := va.Interface().(Marshaler)
	b, err := m.MarshalJSON()
	if err != nil {
		return &MarshalerError{v.Type(), err, "MarshalJSON"}
	}
	printValue(fks, string(b))
	return nil
}

func textMarshalerEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	m, ok := v.Interface().(encoding.TextMarshaler)
	if !ok {
		writeNull(e, fks, v, opts)
		return nil
	}
	b, err := m.MarshalText()
	if err != nil {
		return &MarshalerError{v.Type(), err, "MarshalText"}
	}
	printValue(fks, string(b))
	return nil
}

func addrTextMarshalerEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	va := v.Addr()
	if va.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	m := va.Interface().(encoding.TextMarshaler)
	b, err := m.MarshalText()
	if err != nil {
		return &MarshalerError{v.Type(), err, "MarshalText"}
	}
	printValue(fks, string(b))
	return nil
}

func boolEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	printValue(fks, v.Bool())
	return nil
}

func intEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	printValue(fks, v.Int())
	return nil
}

func uintEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	printValue(fks, v.Uint())
	return nil
}

type floatEncoder int // number of bits

func (bits floatEncoder) encode(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	f := v.Float()
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return &UnsupportedValueError{v, strconv.FormatFloat(f, 'g', -1, int(bits))}
	}
	printValue(fks, v.Float())
	return nil
}

var (
	float32Encoder = (floatEncoder(32)).encode
	float64Encoder = (floatEncoder(64)).encode
)

func stringEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.Type() == numberType {
		numStr := v.String()
		// In Go1.5 the empty string encodes to "0", while this is not a valid number literal
		// we keep compatibility so check validity after this.
		if numStr == "" {
			numStr = "0" // Number's zero-val
		}
		if !isValidNumber(numStr) {
			return fmt.Errorf("json: invalid number literal %q", numStr)
		}
		printValue(fks, v.String())
		return nil
	}
	if opts.quoted {
		printValue(fks, v.String())
		return nil
	}
	printValue(fks, v.String())
	return nil
}

// isValidNumber reports whether s is a valid JSON number literal.
func isValidNumber(s string) bool {
	// This function implements the JSON numbers grammar.
	// See https://tools.ietf.org/html/rfc7159#section-6
	// and https://www.json.org/img/number.png

	if s == "" {
		return false
	}

	// Optional -
	if s[0] == '-' {
		s = s[1:]
		if s == "" {
			return false
		}
	}

	// Digits
	switch {
	default:
		return false

	case s[0] == '0':
		s = s[1:]

	case '1' <= s[0] && s[0] <= '9':
		s = s[1:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// . followed by 1 or more digits.
	if len(s) >= 2 && s[0] == '.' && '0' <= s[1] && s[1] <= '9' {
		s = s[2:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if len(s) >= 2 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		if s[0] == '+' || s[0] == '-' {
			s = s[1:]
			if s == "" {
				return false
			}
		}
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// Make sure we are at the end.
	return s == ""
}

func interfaceEncoder(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	value := v.Elem()
	return valueEncoder(value)(e, fks, value, opts)
}

type structEncoder struct {
	fields structFields
}

type structFields struct {
	list      []field
	nameIndex map[string]int
}

func (se structEncoder) encode(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
FieldLoop:
	for i := range se.fields.list {
		f := &se.fields.list[i]

		// Find the nested struct field by following f.index.
		fv := v
		for _, i := range f.index {
			if fv.Kind() == reflect.Pointer {
				if fv.IsNil() {
					continue FieldLoop
				}
				fv = fv.Elem()
			}
			fv = fv.Field(i)
		}

		if f.omitEmpty && reflectx.IsEmptyValue(fv) {
			continue
		}
		opts.quoted = f.quoted
		if err := f.encoder(e, append(fks, f.name), fv, opts); err != nil {
			return err
		}
	}
	return nil
}

func newStructEncoder(t reflect.Type) encoderFunc {
	se := structEncoder{fields: cachedTypeFields(t)}
	return se.encode
}

type mapEncoder struct {
	elemEnc encoderFunc
}

func (me mapEncoder) encode(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	if e.ptrLevel++; e.ptrLevel > startDetectingCyclesAfter {
		// We're a large number of nested ptrEncoder.encode calls deep;
		// start checking if we've run into a pointer cycle.
		ptr := v.UnsafePointer()
		if _, ok := e.ptrSeen[ptr]; ok {
			return &UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.Type())}
		}
		e.ptrSeen[ptr] = struct{}{}
		defer delete(e.ptrSeen, ptr)
	}

	// Extract and sort the keys.
	sv := make([]reflectWithString, v.Len())
	mi := v.MapRange()
	for i := 0; mi.Next(); i++ {
		sv[i].k = mi.Key()
		sv[i].v = mi.Value()
		if err := sv[i].resolve(); err != nil {
			return fmt.Errorf("json: encoding error for type %q: %q", v.Type().String(), err.Error())
		}
	}
	sort.Slice(sv, func(i, j int) bool { return sv[i].ks < sv[j].ks })

	for _, kv := range sv {
		if err := me.elemEnc(e, append(slices.Clone(fks), kv.ks), kv.v, opts); err != nil {
			return err
		}
	}
	e.ptrLevel--
	return nil
}

func newMapEncoder(t reflect.Type) encoderFunc {
	switch t.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		if !t.Key().Implements(textMarshalerType) {
			return unsupportedTypeEncoder
		}
	}
	me := mapEncoder{typeEncoder(t.Elem())}
	return me.encode
}

func encodeByteSlice(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	printValue(fks, string(v.Bytes()))
	return nil
}

// sliceEncoder just wraps an arrayEncoder, checking to make sure the value isn't nil.
type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se sliceEncoder) encode(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	if e.ptrLevel++; e.ptrLevel > startDetectingCyclesAfter {
		// We're a large number of nested ptrEncoder.encode calls deep;
		// start checking if we've run into a pointer cycle.
		// Here we use a struct to memorize the pointer to the first element of the slice
		// and its length.
		ptr := struct {
			ptr interface{} // always an unsafe.Pointer, but avoids a dependency on package unsafe
			len int
		}{v.UnsafePointer(), v.Len()}
		if _, ok := e.ptrSeen[ptr]; ok {
			return &UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.Type())}
		}
		e.ptrSeen[ptr] = struct{}{}
		defer delete(e.ptrSeen, ptr)
	}
	if err := se.arrayEnc(e, fks, v, opts); err != nil {
		return err
	}
	e.ptrLevel--
	return nil
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		p := reflect.PointerTo(t.Elem())
		if !p.Implements(marshalerType) && !p.Implements(textMarshalerType) {
			return encodeByteSlice
		}
	}
	enc := sliceEncoder{newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae arrayEncoder) encode(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	n := v.Len()
	for i := 0; i < n; i++ {
		if err := ae.elemEnc(e, append(slices.Clone(fks), convx.ToString(i)), v.Index(i), opts); err != nil {
			return err
		}
	}
	return nil
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := arrayEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe ptrEncoder) encode(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.IsNil() {
		writeNull(e, fks, v, opts)
		return nil
	}
	if e.ptrLevel++; e.ptrLevel > startDetectingCyclesAfter {
		// We're a large number of nested ptrEncoder.encode calls deep;
		// start checking if we've run into a pointer cycle.
		ptr := v.Interface()
		if _, ok := e.ptrSeen[ptr]; ok {
			return &UnsupportedValueError{v, fmt.Sprintf("encountered a cycle via %s", v.Type())}
		}
		e.ptrSeen[ptr] = struct{}{}
		defer delete(e.ptrSeen, ptr)
	}
	if err := pe.elemEnc(e, fks, v.Elem(), opts); err != nil {
		return err
	}
	e.ptrLevel--
	return nil
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := ptrEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type condAddrEncoder struct {
	canAddrEnc, elseEnc encoderFunc
}

func (ce condAddrEncoder) encode(e *encodeState, fks []string, v reflect.Value, opts encOpts) error {
	if v.CanAddr() {
		return ce.canAddrEnc(e, fks, v, opts)
	}
	return ce.elseEnc(e, fks, v, opts)
}

// newCondAddrEncoder returns an encoder that checks whether its value
// CanAddr and delegates to canAddrEnc if so, else to elseEnc.
func newCondAddrEncoder(canAddrEnc, elseEnc encoderFunc) encoderFunc {
	enc := condAddrEncoder{canAddrEnc: canAddrEnc, elseEnc: elseEnc}
	return enc.encode
}

func unsupportedTypeEncoder(e *encodeState, fks []string, v reflect.Value, _ encOpts) error {
	return &UnsupportedTypeError{v.Type()}
}
