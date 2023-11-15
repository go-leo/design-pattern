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
)

type ClonerFunc func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error

// emptyValueCloner 是将一个空值复制到目标值中。
// 它首先检查目标值是否有效，如果无效则直接返回。
// 然后，根据目标值的类型，将其设置为相应类型的零值或忽略 nil 值。
func emptyValueCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	scanner, tv := IndirectValue(tgtVal)
	if scanner != nil {
		return scanner.Scan(nil)
	}
	switch tv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice:
		tv.Set(reflect.Zero(tv.Type()))
		return nil
	case reflect.Interface:
		tv.Set(reflect.Zero(tv.Type()))
		return nil
	default:
		// otherwise, ignore nil for primitives/string
		return nil
	}
}

// boolCloner 将源值 srcVal 的布尔值复制到目标值 tgtVal 中。
// 它首先检查目标值是否有效，如果无效则直接返回。
// 然后，根据目标值的类型，将源值的布尔值设置到目标值中，或者根据情况调用 hookCloner 函数处理。
func boolCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	b := srcVal.Bool()
	scanner, tv := IndirectValue(tgtVal)
	if scanner != nil {
		return scanner.Scan(b)
	}
	switch tv.Kind() {
	case reflect.Bool:
		tv.SetBool(b)
	case reflect.Interface:
		if tv.NumMethod() == 0 {
			tv.Set(reflect.ValueOf(b))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	return nil
}

func intCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	i := srcVal.Int()
	scanner, tv := IndirectValue(tgtVal)
	if scanner != nil {
		return scanner.Scan(i)
	}
	switch tv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if tv.OverflowInt(i) {
			return &OverflowError{
				FullKeys:    fks,
				TargetValue: tgtVal,
				Value:       strconv.FormatInt(i, 10),
			}
		}
		tv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if i < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatInt(i, 10)}
		}
		u := uint64(i)
		if tv.OverflowUint(u) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatInt(i, 10)}
		}
		tv.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		f := float64(i)
		if tv.OverflowFloat(f) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatInt(i, 10)}
		}
		tv.SetFloat(f)
		return nil
	case reflect.Interface:
		if tv.NumMethod() == 0 {
			tv.Set(reflect.ValueOf(i))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Struct:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func uintCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	u := srcVal.Uint()
	scanner, tv := IndirectValue(tgtVal)
	if scanner != nil {
		return scanner.Scan(u)
	}
	switch tv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if u > uint64(math.MaxInt64) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatUint(u, 10)}
		}
		i := int64(u)
		if tv.OverflowInt(i) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatUint(u, 10)}
		}
		tv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if tv.OverflowUint(u) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatUint(u, 10)}
		}
		tv.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		f := float64(u)
		if tv.OverflowFloat(f) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatUint(u, 10)}
		}
		tv.SetFloat(f)
		return nil
	case reflect.Interface:
		if tv.NumMethod() == 0 {
			tv.Set(reflect.ValueOf(u))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

var (
	float32Cloner = (floatCloner(32)).encode
	float64Cloner = (floatCloner(64)).encode
)

type floatCloner int // number of bits

func (bits floatCloner) encode(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	f := srcVal.Float()
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return &UnsupportedValueError{Value: srcVal, Str: strconv.FormatFloat(f, 'g', -1, int(bits))}
	}
	scanner, tv := IndirectValue(tgtVal)
	if scanner != nil {
		return scanner.Scan(f)
	}
	switch tv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
			return &OverflowError{
				FullKeys:    fks,
				TargetValue: tgtVal,
				Value:       strconv.FormatFloat(f, 'g', -1, 64),
			}
		}
		i := int64(f)
		if tv.OverflowInt(i) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatFloat(f, 'g', -1, int(bits))}
		}
		tv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if f > float64(math.MaxUint64) {
			return &OverflowError{
				FullKeys:    fks,
				TargetValue: tgtVal,
				Value:       strconv.FormatFloat(f, 'g', -1, 64),
			}
		}
		if f < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatFloat(f, 'g', -1, int(bits))}
		}
		u := uint64(f)
		if tv.OverflowUint(u) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatFloat(f, 'g', -1, int(bits))}
		}
		tv.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		if tv.OverflowFloat(f) {
			return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatFloat(f, 'g', -1, int(bits))}
		}
		tv.SetFloat(f)
		return nil
	case reflect.Interface:
		if tv.NumMethod() == 0 {
			tv.Set(reflect.ValueOf(f))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func stringCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	s := srcVal.String()
	scanner, tv := IndirectValue(tgtVal)
	if scanner != nil {
		return scanner.Scan(s)
	}
	switch tv.Kind() {
	case reflect.Slice:
		if tv.Type().Elem().Kind() != reflect.Uint8 {
			return &CloneError{Value: "string", Type: tv.Type()}
		}
		tv.SetBytes([]byte(s))
	case reflect.String:
		tv.SetString(s)
	case reflect.Interface:
		if tv.NumMethod() == 0 {
			tv.Set(reflect.ValueOf(s))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	return nil
}

func interfaceCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	if srcVal.IsNil() {
		return emptyValueCloner(e, fks, tgtVal, srcVal, opts)
	}
	srcVal = srcVal.Elem()
	return valueCloner(srcVal, opts)(e, fks, tgtVal, srcVal, opts)
}

func structCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	scanner, tv := IndirectValue(tgtVal)
	if scanner != nil {
		return scanner.Scan(srcVal.Interface())
	}
	switch tv.Kind() {
	case reflect.Bool:
		return struct2BoolCloner(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return struct2IntCloner(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return struct2UintCloner(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Float32, reflect.Float64:
		return struct2FloatCloner(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.String:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Slice:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Struct:
		return struct2StructCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Interface:
		return struct2InterfaceCloner(e, fks, tgtVal, srcVal, opts, tv)
	case reflect.Map:
		// TODO Map cloner
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Array:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func struct2BoolCloner(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, tv reflect.Value) error {
	var b bool
	switch srcVal.Type() {
	case sqlNullBoolType:
		b = srcVal.FieldByName("Bool").Bool()
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	tv.SetBool(b)
	return nil
}

func struct2IntCloner(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, tv reflect.Value) error {
	var i int64
	switch srcVal.Type() {
	case sqlNullByteType:
		i = int64(srcVal.FieldByName("Byte").Uint())
	case sqlNullInt16Type:
		i = srcVal.FieldByName("Int16").Int()
	case sqlNullInt32Type:
		i = srcVal.FieldByName("Int32").Int()
	case sqlNullInt64Type:
		i = srcVal.FieldByName("Int64").Int()
	case sqlNullFloat64Type:
		f := srcVal.FieldByName("Float64").Float()
		if f > float64(math.MaxInt64) || f < float64(math.MinInt64) {
			return &OverflowError{
				FullKeys:    fks,
				TargetValue: tgtVal,
				Value:       strconv.FormatFloat(f, 'g', -1, 64),
			}
		}
		i = int64(f)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	if tv.OverflowInt(i) {
		return &OverflowError{
			FullKeys:    fks,
			TargetValue: tgtVal,
			Value:       strconv.FormatInt(i, 10),
		}
	}
	tv.SetInt(i)
	return nil
}

func struct2UintCloner(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, tv reflect.Value) error {
	var u uint64
	switch srcVal.Type() {
	case sqlNullByteType:
		u = srcVal.FieldByName("Byte").Uint()
	case sqlNullInt16Type:
		i := srcVal.FieldByName("Int16").Int()
		if i < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatInt(i, 10)}
		}
		u = uint64(i)
	case sqlNullInt32Type:
		i := srcVal.FieldByName("Int32").Int()
		if i < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatInt(i, 10)}
		}
		u = uint64(i)
	case sqlNullInt64Type:
		i := srcVal.FieldByName("Int64").Int()
		if i < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatInt(i, 10)}
		}
		u = uint64(i)
	case sqlNullFloat64Type:
		f := srcVal.FieldByName("Float64").Float()
		if f > float64(math.MaxUint64) {
			return &OverflowError{
				FullKeys:    fks,
				TargetValue: tgtVal,
				Value:       strconv.FormatFloat(f, 'g', -1, 64),
			}
		}
		if f < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatFloat(f, 'g', -1, 64)}
		}
		u = uint64(f)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	if tv.OverflowUint(u) {
		return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatUint(u, 10)}
	}
	tv.SetUint(u)
	return nil
}

func struct2FloatCloner(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, tv reflect.Value) error {
	var f float64
	switch srcVal.Type() {
	case sqlNullByteType:
		f = float64(srcVal.FieldByName("Byte").Uint())
	case sqlNullInt16Type:
		f = float64(srcVal.FieldByName("Int16").Int())
	case sqlNullInt32Type:
		f = float64(srcVal.FieldByName("Int32").Int())
	case sqlNullInt64Type:
		f = float64(srcVal.FieldByName("Int64").Int())
	case sqlNullFloat64Type:
		f = srcVal.FieldByName("Float64").Float()
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	if tv.OverflowFloat(f) {
		return &OverflowError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatFloat(f, 'g', -1, 64)}
	}
	tv.SetFloat(f)
	return nil
}

func struct2InterfaceCloner(e *cloneContext, fks []string, tgtVal reflect.Value, srcVal reflect.Value, opts *options, tv reflect.Value) error {
	if tv.NumMethod() == 0 {
		// 创建一个新的空对象
		v := reflect.New(srcVal.Type())
		clonedVal := v.Elem()
		if err := struct2StructCloner(e, fks, clonedVal, srcVal, opts); err != nil {
			return err
		}
		// 设置到目标对象
		tv.Set(v)
		return nil
	}
	return hookCloner(e, fks, tgtVal, srcVal, opts)
}

func struct2StructCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	tgtType := tgtVal.Type()
	tgtFields := cachedTypeFields(tgtType, opts, false)
	srcType := srcVal.Type()
	srcFields := cachedTypeFields(srcType, opts, true)
	// 复制字段, 循环src字段
FieldLoop:
	for srcKey, srcIdx := range srcFields.nameIndex {
		srcField := srcFields.list[srcIdx]
		srcFieldVal := srcVal

		// 查找src字段值
		for _, i := range srcField.index {
			if srcFieldVal.Kind() == reflect.Pointer {
				if srcFieldVal.IsNil() {
					continue FieldLoop
				}
				srcFieldVal = srcFieldVal.Elem()
			}
			srcFieldVal = srcFieldVal.Field(i)
		}

		// 查找tgt字段
		var tgtField *field
		tgtIdx, ok := tgtFields.nameIndex[srcKey]
		if ok {
			// 找到了一个完全匹配的字段名称
			tgtField = &tgtFields.list[tgtIdx]
		} else {
			// 代码回退到了一种更为耗时的线性搜索方法，该方法在进行字段名称匹配时不考虑大小写
			for tgtKey, tgtIdx := range tgtFields.nameIndex {
				if opts.EqualFold(tgtKey, srcKey) {
					tgtField = &tgtFields.list[tgtIdx]
					break
				}
			}
		}
		if tgtField == nil {
			// 没有找到目标，则跳过
			continue FieldLoop
		}

		// 查找tgt字段值
		tgtFieldVal := tgtVal
		for _, i := range tgtField.index {
			if tgtFieldVal.Kind() == reflect.Pointer {
				if tgtFieldVal.IsNil() {
					if !tgtFieldVal.CanSet() {
						return fmt.Errorf("prototype: cannot set embedded pointer to unexported struct: %v", tgtFieldVal.Type().Elem())
					}
					tgtFieldVal.Set(reflect.New(tgtFieldVal.Type().Elem()))
				}
				tgtFieldVal = tgtFieldVal.Elem()
			}
			tgtFieldVal = tgtFieldVal.Field(i)
		}

		// 如果src字段是空值，则克隆空值
		if reflectx.IsEmptyValue(srcFieldVal) {
			if err := emptyValueCloner(e, append(slices.Clone(fks), srcKey), tgtFieldVal, srcFieldVal, opts); err != nil {
				return err
			}
			continue
		}

		// 克隆src字段到tgt字段
		if err := srcField.clonerFunc(e, append(slices.Clone(fks), srcKey), tgtFieldVal, srcFieldVal, opts); err != nil {
			return err
		}
	}
	return nil
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

type mapEncoder struct {
	elemEnc ClonerFunc
}

func (me mapEncoder) encode(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return emptyValueCloner(e, fks, tgtVal, srcVal, opts)
	}
	if e.forward(); e.isTooDeep() {
		// We're a large number of nested ptrCloner.encode calls deep;
		// start checking if we've run into a pointer cycle.
		ptr := srcVal.UnsafePointer()
		if e.isSeen(ptr) {
			return &UnsupportedValueError{Value: srcVal, Str: fmt.Sprintf("encountered a cycle via %s", srcVal.Type())}
		}
		e.remember(ptr)
		defer e.forget(ptr)
	}

	// Extract and sort the keys.
	sv := make([]reflectWithString, srcVal.Len())
	mi := srcVal.MapRange()
	for i := 0; mi.Next(); i++ {
		sv[i].k = mi.Key()
		sv[i].v = mi.Value()
		if err := sv[i].resolve(); err != nil {
			return fmt.Errorf("json: encoding error for type %q: %q", srcVal.Type().String(), err.Error())
		}
	}
	sort.Slice(sv, func(i, j int) bool { return sv[i].ks < sv[j].ks })

	for _, kv := range sv {
		if err := me.elemEnc(e, append(slices.Clone(fks), kv.ks), tgtVal, kv.v, opts); err != nil {
			return err
		}
	}
	e.back()
	return nil
}

func newMapCloner(srcType reflect.Type, opts *options) ClonerFunc {
	switch srcType.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		if !srcType.Key().Implements(textMarshalerType) {
			return unsupportedTypeCloner
		}
	}
	me := mapEncoder{elemEnc: typeCloner(srcType.Elem(), opts)}
	return me.encode
}

// sliceCloner just wraps an arrayCloner, checking to make sure the value isn't nil.
type sliceCloner struct {
	arrayCloner ClonerFunc
}

func newSliceCloner(srcType reflect.Type, opts *options) ClonerFunc {
	enc := sliceCloner{arrayCloner: newArrayCloner(srcType, opts)}
	return enc.clone
}

func (se sliceCloner) clone(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	if srcVal.IsNil() {
		return emptyValueCloner(e, fks, tgtVal, srcVal, opts)
	}
	if e.forward(); e.isTooDeep() {
		// We're a large number of nested ptrCloner.encode calls deep;
		// start checking if we've run into a pointer cycle.
		// Here we use a struct to memorize the pointer to the first element of the slice
		// and its length.
		ptr := struct {
			ptr interface{} // always an unsafe.Pointer, but avoids a dependency on package unsafe
			len int
		}{srcVal.UnsafePointer(), srcVal.Len()}
		if e.isSeen(ptr) {
			return &UnsupportedValueError{Value: srcVal, Str: fmt.Sprintf("encountered a cycle via %s", srcVal.Type())}
		}
		e.remember(ptr)
		defer e.forget(ptr)
	}
	if err := se.arrayCloner(e, fks, tgtVal, srcVal, opts); err != nil {
		return err
	}
	e.back()
	return nil
}

type arrayCloner struct {
	elemEnc ClonerFunc
}

func newArrayCloner(srcType reflect.Type, opts *options) ClonerFunc {
	enc := arrayCloner{elemEnc: typeCloner(srcType.Elem(), opts)}
	return enc.encode
}

func (ae arrayCloner) encode(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	tv := indirect(tgtVal, false)
	switch tv.Kind() {
	case reflect.Array, reflect.Slice:
		srcLen := srcVal.Len()
		if tv.Kind() == reflect.Slice {
			tv.Set(reflect.MakeSlice(tv.Type(), srcLen, srcLen))
			tv.SetLen(0)
		}
		for i := 0; i < srcVal.Len(); i++ {
			if tv.Kind() == reflect.Slice {
				tv.SetLen(i + 1)
			}
			if i >= tv.Len() {
				// Ran out of fixed array: skip.
				continue
			}
			tgtItem := tv.Index(i)
			srcItem := srcVal.Index(i)
			if err := ae.elemEnc(e, append(slices.Clone(fks), convx.ToString(i)), tgtItem, srcItem, opts); err != nil {
				return err
			}
			tv.Index(i).Set(tgtItem)
		}
		return nil
	case reflect.Interface:
		if tv.NumMethod() > 0 {
			return hookCloner(e, fks, tgtVal, srcVal, opts)
		}
		return ae.anySliceClone(e, fks, tv, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func (ae arrayCloner) anySliceClone(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcLen := srcVal.Len()
	// 获取src的反射类型
	srcType := srcVal.Type()
	// 创建一个与src类型相同的切片类型
	tgtType := reflect.SliceOf(srcType.Elem())
	// 使用反射创建一个切片
	tgtSlice := reflect.MakeSlice(tgtType, srcLen, srcLen)
	// 将src的元素逐个拷贝到tgt
	for i := 0; i < srcLen; i++ {
		tgtItem := tgtSlice.Index(i)
		srcItem := srcVal.Index(i)
		if err := ae.elemEnc(e, append(slices.Clone(fks), convx.ToString(i)), tgtItem, srcItem, opts); err != nil {
			return err
		}
		tgtSlice.Index(i).Set(tgtItem)
	}
	// 设置tgtVal
	tgtVal.Set(tgtSlice)
	return nil
}

type ptrCloner struct {
	elemEnc ClonerFunc
}

func newPtrCloner(srcType reflect.Type, opts *options) ClonerFunc {
	enc := ptrCloner{elemEnc: typeCloner(srcType.Elem(), opts)}
	return enc.encode
}

func (pe ptrCloner) encode(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return emptyValueCloner(e, fks, tgtVal, srcVal, opts)
	}
	if e.forward(); e.isTooDeep() {
		// We're a large number of nested ptrCloner.encode calls deep;
		// start checking if we've run into a pointer cycle.
		ptr := srcVal.Interface()
		if e.isSeen(ptr) {
			return &UnsupportedValueError{Value: srcVal, Str: fmt.Sprintf("encountered a cycle via %s", srcVal.Type())}
		}
		e.remember(ptr)
		defer e.forget(ptr)
	}
	if err := pe.elemEnc(e, fks, tgtVal, srcVal.Elem(), opts); err != nil {
		return err
	}
	e.back()
	return nil
}

func unsupportedTypeCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	return hookCloner(e, fks, tgtVal, srcVal, opts)
}

func hookCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	valueHooks, ok := opts.ValueHook[srcVal]
	if ok {
		hook, ok := valueHooks[tgtVal]
		if ok {
			return hook(fks, tgtVal, srcVal)
		}
	}
	typeHooks, ok := opts.TypeHooks[srcVal.Type()]
	if ok {
		hook, ok := typeHooks[tgtVal.Type()]
		if ok {
			return hook(fks, tgtVal, srcVal)
		}
	}
	kindHooks, ok := opts.KindHooks[srcVal.Kind()]
	if ok {
		hook, ok := kindHooks[tgtVal.Kind()]
		if ok {
			return hook(fks, tgtVal, srcVal)
		}
	}
	return &UnsupportedTypeError{Type: srcVal.Type()}
}
