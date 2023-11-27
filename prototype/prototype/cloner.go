package prototype

import (
	"encoding"
	"errors"
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
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatInt(i, 10)}
		}
		tv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if i < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatInt(i, 10)}
		}
		u := uint64(i)
		if tv.OverflowUint(u) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatInt(i, 10)}
		}
		tv.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		f := float64(i)
		if tv.OverflowFloat(f) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatInt(i, 10)}
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
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatUint(u, 10)}
		}
		i := int64(u)
		if tv.OverflowInt(i) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatUint(u, 10)}
		}
		tv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if tv.OverflowUint(u) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatUint(u, 10)}
		}
		tv.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		f := float64(u)
		if tv.OverflowFloat(f) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatUint(u, 10)}
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
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatFloat(f, 'g', -1, 64)}
		}
		i := int64(f)
		if tv.OverflowInt(i) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatFloat(f, 'g', -1, int(bits))}
		}
		tv.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if f > float64(math.MaxUint64) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatFloat(f, 'g', -1, 64)}
		}
		if f < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatFloat(f, 'g', -1, int(bits))}
		}
		u := uint64(f)
		if tv.OverflowUint(u) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatFloat(f, 'g', -1, int(bits))}
		}
		tv.SetUint(u)
		return nil
	case reflect.Float32, reflect.Float64:
		if tv.OverflowFloat(f) {
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatFloat(f, 'g', -1, int(bits))}
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
	if srcVal.IsNil() {
		return nil
	}
	if !tgtVal.IsValid() {
		return nil
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
		return struct2StructCloner(e, fks, tv, srcVal, opts)
	case reflect.Interface:
		if tv.NumMethod() == 0 {
			// 创建一个新的空对象
			v := reflect.New(srcVal.Type())
			if err := struct2StructCloner(e, fks, v.Elem(), srcVal, opts); err != nil {
				return err
			}
			// 设置到目标对象
			tv.Set(v)
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Map:
		t := tv.Type()
		switch t.Key().Kind() {
		case reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if tv.IsNil() {
				tv.Set(reflect.MakeMap(t))
			}
			return struct2MapCloner(e, fks, tv, srcVal, opts)
		default:
			if reflect.PointerTo(t.Key()).Implements(textUnmarshalerType) {
				if tv.IsNil() {
					tv.Set(reflect.MakeMap(t))
				}
				return struct2MapCloner(e, fks, tv, srcVal, opts)
			}
		}
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
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatFloat(f, 'g', -1, 64)}
		}
		i = int64(f)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	if tv.OverflowInt(i) {
		return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatInt(i, 10)}
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
			return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatFloat(f, 'g', -1, 64)}
		}
		if f < 0 {
			return &NegativeNumberError{FullKeys: fks, TargetValue: tgtVal, Value: strconv.FormatFloat(f, 'g', -1, 64)}
		}
		u = uint64(f)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
	if tv.OverflowUint(u) {
		return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatUint(u, 10)}
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
		return &OverflowError{FullKeys: fks, TargetType: tgtVal.Type(), Value: strconv.FormatFloat(f, 'g', -1, 64)}
	}
	tv.SetFloat(f)
	return nil
}

func struct2StructCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	tgtType := tgtVal.Type()
	tgtFields := cachedTypeFields(tgtType, opts, opts.TargetTagKey)
	srcType := srcVal.Type()
	srcFields := cachedTypeFields(srcType, opts, opts.SourceTagKey)
	if err := struct2StructDominantFieldCloner(e, fks, tgtVal, srcVal, tgtType, srcType, tgtFields, srcFields, opts); err != nil {
		return err
	}
	if err := struct2StructRecessivesFieldCloner(e, fks, tgtVal, srcVal, tgtType, srcType, tgtFields, srcFields, opts); err != nil {
		return err
	}
	return nil
}

func struct2StructDominantFieldCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, tgtType, srcType reflect.Type, tgtFields, srcFields structFields, opts *options) error {
	// 复制字段, 循环src字段
	for srcName, srcIdx := range srcFields.dominantsNameIndex {
		srcDominantField := srcFields.dominants[srcIdx]
		// 查找src字段值
		srcDominantFieldVal, ok := findValue(srcVal, srcDominantField)
		if !ok {
			continue
		}

		// 查找 tgt 主要字段
		tgtDominantField, ok := findDominantField(tgtFields, opts, srcName)
		if !ok {
			// 没有找到目标，则跳过
			continue
		}

		// 查找tgt字段值
		tgtDominantFieldVal, err := findSettableValue(tgtVal, tgtDominantField)
		if err != nil {
			return err
		}

		// 如果src字段是空值，则克隆空值
		if reflectx.IsEmptyValue(srcDominantFieldVal) {
			if err := emptyValueCloner(e, append(slices.Clone(fks), srcName), tgtDominantFieldVal, srcDominantFieldVal, opts); err != nil {
				return err
			}
			continue
		}

		// 克隆src字段到tgt字段
		if err := srcDominantField.clonerFunc(e, append(slices.Clone(fks), srcName), tgtDominantFieldVal, srcDominantFieldVal, opts); err != nil {
			return err
		}
	}
	return nil
}

func struct2StructRecessivesFieldCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, tgtType, srcType reflect.Type, tgtFields, srcFields structFields, opts *options) error {
	// 复制字段, 循环src字段
	for srcKey, srcIdxs := range srcFields.recessivesNameIndex {
		srcRecessiveFieldValMap := make(map[string]reflect.Value)
		for _, srcIdx := range srcIdxs {
			srcRecessiveField := srcFields.recessives[srcIdx]
			// 查找src字段值
			srcRecessiveFieldVal, ok := findValue(srcVal, srcRecessiveField)
			if !ok {
				continue
			}
			srcRecessiveFieldValMap[srcRecessiveField.fullName] = srcRecessiveFieldVal
		}
		if len(srcRecessiveFieldValMap) <= 0 {
			continue
		}

		tgtRecessiveFields, ok := findRecessiveField(tgtFields, opts, srcKey)
		if !ok {
			continue
		}
		if len(tgtRecessiveFields) <= 0 {
			continue
		}

		tgtRecessiveFieldValMap := make(map[string]reflect.Value)
		for _, recessiveField := range tgtRecessiveFields {
			// 查找tgt字段值
			tgtFieldVal, err := findSettableValue(tgtVal, recessiveField)
			if err != nil {
				return err
			}
			tgtRecessiveFieldValMap[recessiveField.fullName] = tgtFieldVal
		}

		for fullName, srcRecessiveFieldVal := range srcRecessiveFieldValMap {
			tgtRecessiveFieldVal, ok := tgtRecessiveFieldValMap[fullName]
			if !ok {
				continue
			}
			// 如果src字段是空值，则克隆空值
			if reflectx.IsEmptyValue(srcRecessiveFieldVal) {
				if err := emptyValueCloner(e, append(slices.Clone(fks), srcKey), tgtRecessiveFieldVal, srcRecessiveFieldVal, opts); err != nil {
					return err
				}
				continue
			}

			// 克隆src字段到tgt字段
			srcRecessiveField := srcFields.recessives[srcFields.recessivesFullNameIndex[fullName]]
			if err := srcRecessiveField.clonerFunc(e, append(slices.Clone(fks), srcKey), tgtRecessiveFieldVal, srcRecessiveFieldVal, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func struct2MapCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcType := srcVal.Type()
	srcFields := cachedTypeFields(srcType, opts, opts.SourceTagKey)
	_ = srcFields
	return nil
}

func struct2MapDominantFieldCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, srcFields structFields, t reflect.Type, opts *options) error {
	// 复制字段, 循环src字段
	var mapElem reflect.Value
	elemType := t.Elem()
	for srcName, srcIdx := range srcFields.dominantsNameIndex {
		srcDominantField := srcFields.dominants[srcIdx]
		// 查找src字段值
		srcDominantFieldVal, ok := findValue(srcVal, srcDominantField)
		if !ok {
			continue
		}

		if !mapElem.IsValid() {
			mapElem = reflect.New(elemType).Elem()
		} else {
			mapElem.Set(reflect.Zero(elemType))
		}
		subv := mapElem

		// 如果src字段是空值，则克隆空值
		if reflectx.IsEmptyValue(srcDominantFieldVal) {
			if err := emptyValueCloner(e, append(slices.Clone(fks), srcName), subv, srcDominantFieldVal, opts); err != nil {
				return err
			}
		} else {
			// 克隆src字段到tgt字段
			if err := srcDominantField.clonerFunc(e, append(slices.Clone(fks), srcName), subv, srcDominantFieldVal, opts); err != nil {
				return err
			}
		}

		if keyVal.IsValid() {
			tgtVal.SetMapIndex(keyVal, subv)
		}
	}
	return nil
}

func struct2MapDominantFieldCloners(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, srcFields structFields, t reflect.Type, key string, opts *options) (reflect.Value, error) {
	keyType := t.Key()
	switch {
	case reflect.PointerTo(keyType).Implements(textUnmarshalerType):
		keyVal := reflect.New(keyType)
		if textUnmarshaler, ok := keyVal.Interface().(encoding.TextUnmarshaler); ok {
			if err := textUnmarshaler.UnmarshalText([]byte(key)); err != nil {
				return reflect.Value{}, err
			}
		}
		return keyVal.Elem(), nil
	case keyType.Kind() == reflect.String:
		return reflect.ValueOf(key).Convert(keyType), nil
	default:
		switch keyType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			if reflect.Zero(keyType).OverflowInt(n) {
				return reflect.Value{}, &OverflowError{FullKeys: fks, TargetType: keyType, Value: key}
			}
			return reflect.ValueOf(n).Convert(keyType), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			n, err := strconv.ParseUint(key, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			if reflect.Zero(keyType).OverflowUint(n) {
				return reflect.Value{}, &OverflowError{FullKeys: fks, TargetType: keyType, Value: key}
			}
			return reflect.ValueOf(n).Convert(keyType), nil
		}
	}
	return reflect.Value{}, errors.New("prototype: Unexpected key type") // should never occur
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

func mapCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcType := srcVal.Type()
	switch srcType.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		if !srcType.Key().Implements(textMarshalerType) {
			return hookCloner(e, fks, tgtVal, srcVal, opts)
		}
	}
	if srcVal.IsNil() {
		return nil
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

	cloner := typeCloner(srcType.Elem(), opts)
	for _, kv := range sv {
		if err := cloner(e, append(slices.Clone(fks), kv.ks), tgtVal, kv.v, opts); err != nil {
			return err
		}
	}
	e.back()
	return nil
}

// sliceCloner just wraps an arrayCloner, checking to make sure the value isn't nil.
func sliceCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}
	if !tgtVal.IsValid() {
		return nil
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
	if err := arrayCloner(e, fks, tgtVal, srcVal, opts); err != nil {
		return err
	}
	e.back()
	return nil
}

func arrayCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
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
		elemEnc := typeCloner(tgtVal.Type().Elem(), opts)
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
			if err := elemEnc(e, append(slices.Clone(fks), convx.ToString(i)), tgtItem, srcItem, opts); err != nil {
				return err
			}
			tv.Index(i).Set(tgtItem)
		}
		return nil
	case reflect.Interface:
		if tv.NumMethod() > 0 {
			return hookCloner(e, fks, tgtVal, srcVal, opts)
		}
		return anySliceClone(e, fks, tv, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func anySliceClone(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	elemEnc := typeCloner(tgtVal.Type().Elem(), opts)
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
		if err := elemEnc(e, append(slices.Clone(fks), convx.ToString(i)), tgtItem, srcItem, opts); err != nil {
			return err
		}
		tgtSlice.Index(i).Set(tgtItem)
	}
	// 设置tgtVal
	tgtVal.Set(tgtSlice)
	return nil
}

func ptrCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
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
	elemEnc := typeCloner(srcVal.Type().Elem(), opts)
	if err := elemEnc(e, fks, tgtVal, srcVal.Elem(), opts); err != nil {
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
