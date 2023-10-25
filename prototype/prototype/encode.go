package prototype

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

func Clone(tgt any, src any) error {
	tgtVal := reflect.ValueOf(tgt)
	if tgtVal.Kind() != reflect.Pointer || tgtVal.IsNil() {
		return &InvalidCloneError{Type: reflect.TypeOf(tgt)}
	}
	srcVal := reflect.ValueOf(src)

	e := newCloneState()
	defer cloneStatePool.Put(e)

	return clone(e, tgtVal, srcVal, options{})
}

func clone(e *cloneState, tgtVal, srcVal reflect.Value, opts options) error {
	return valueEncoder(srcVal)(e, []string{}, srcVal, opts)
}

type options struct{}

var encoderCache sync.Map // map[reflect.Type]encoderFunc

func valueEncoder(srvVal reflect.Value) encoderFunc {
	if !srvVal.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(srvVal.Type())
}

func typeEncoder(srcType reflect.Type) encoderFunc {
	if fi, ok := encoderCache.Load(srcType); ok {
		return fi.(encoderFunc)
	}

	// To deal with recursive types, populate the map with an
	// indirect func before we build it. This type waits on the
	// real func (f) to be ready and then calls it. This indirect
	// func is only used for recursive types.
	var (
		wg sync.WaitGroup
		f  encoderFunc
	)
	wg.Add(1)
	fi, loaded := encoderCache.LoadOrStore(srcType, encoderFunc(func(e *cloneState, fks []string, srcVal reflect.Value, opts options) error {
		wg.Wait()
		return f(e, fks, srcVal, opts)
	}))
	if loaded {
		return fi.(encoderFunc)
	}

	// Compute the real encoder and replace the indirect func with it.
	f = newTypeEncoder(srcType, true)
	wg.Done()
	encoderCache.Store(srcType, f)
	return f
}

// newTypeEncoder constructs an encoderFunc for a type.
// The returned encoder only checks CanAddr when allowAddr is true.
func newTypeEncoder(srcType reflect.Type, allowAddr bool) encoderFunc {
	// If we have a non-pointer value whose type implements
	// Marshaler with a value receiver, then we're better off taking
	// the address of the value - otherwise we end up with an
	// allocation as we cast the value to an interface.
	if srcType.Kind() != reflect.Pointer && allowAddr && reflect.PointerTo(srcType).Implements(textMarshalerType) {
		return newCondAddrEncoder(addrTextMarshalerEncoder, newTypeEncoder(srcType, false))
	}
	if srcType.Implements(textMarshalerType) {
		return textMarshalerEncoder
	}

	switch srcType.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32:
		return float32Encoder
	case reflect.Float64:
		return float64Encoder
	case reflect.String:
		return stringEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct:
		return newStructEncoder(srcType)
	case reflect.Map:
		return newMapEncoder(srcType)
	case reflect.Slice:
		return newSliceEncoder(srcType)
	case reflect.Array:
		return newArrayEncoder(srcType)
	case reflect.Pointer:
		return newPtrEncoder(srcType)
	default:
		return unsupportedTypeEncoder
	}
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			return false
		}
	}
	return true
}

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
}

type reflectWithString struct {
	k  reflect.Value
	v  reflect.Value
	ks string
}

func (w *reflectWithString) resolve() error {
	if w.k.Kind() == reflect.String {
		w.ks = w.k.String()
		return nil
	}
	if tm, ok := w.k.Interface().(encoding.TextMarshaler); ok {
		if w.k.Kind() == reflect.Pointer && w.k.IsNil() {
			return nil
		}
		buf, err := tm.MarshalText()
		w.ks = string(buf)
		return err
	}
	switch w.k.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.ks = strconv.FormatInt(w.k.Int(), 10)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		w.ks = strconv.FormatUint(w.k.Uint(), 10)
		return nil
	}
	panic("unexpected map key type")
}