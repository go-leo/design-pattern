package prototype

import (
	"encoding"
	"errors"
	"fmt"
	"github.com/go-leo/gox/convx"
	"golang.org/x/exp/slices"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// clonerFunc 通用克隆方法
type clonerFunc func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error

func clonerByValue(srcVal reflect.Value, opts *options) clonerFunc {
	if !srcVal.IsValid() {
		return func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
			return nil
		}
	}
	return clonerByType(srcVal.Type(), true, opts)
}

// newCondAddrEncoder returns an encoder that checks whether its value
// CanAddr and delegates to canAddrEnc if so, else to elseEnc.
func newCondAddrEncoder(canAddrEnc, elseEnc clonerFunc) clonerFunc {
	return func(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
		if srcVal.CanAddr() {
			return canAddrEnc(e, fks, tgtVal, srcVal, opts)
		} else {
			return elseEnc(e, fks, tgtVal, srcVal, opts)
		}
	}
}

func clonerToCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.Kind() == reflect.Pointer && srcVal.IsNil() {
		return nil
	}
	if cloner, ok := srcVal.Interface().(ClonerTo); ok {
		return cloner.CloneTo(tgtVal)
	}
	return nil
}

func addrClonerToCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcAddr := srcVal.Addr()
	if srcAddr.IsNil() {
		return nil
	}
	if cloner, ok := srcAddr.Interface().(ClonerTo); ok {
		return cloner.CloneTo(tgtVal)
	}
	return nil
}

// clonerByType 基于 reflect.Type 获取 clonerFunc
func clonerByType(srcType reflect.Type, allowAddr bool, opts *options) clonerFunc {
	if srcType.Kind() != reflect.Pointer && allowAddr && reflect.PointerTo(srcType).Implements(clonerToType) {
		return newCondAddrEncoder(addrClonerToCloner, clonerByType(srcType, false, opts))
	}
	if srcType.Implements(clonerToType) {
		return clonerToCloner
	}
	switch srcType.Kind() {
	case reflect.Bool:
		return boolCloner
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intCloner
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintCloner
	case reflect.Float32, reflect.Float64:
		return floatCloner
	case reflect.String:
		return stringCloner
	case reflect.Struct:
		return structCloner
	case reflect.Map:
		return mapCloner
	case reflect.Slice:
		return sliceCloner
	case reflect.Array:
		return arrayCloner
	case reflect.Interface:
		return interfaceCloner
	case reflect.Pointer:
		return ptrCloner
	default:
		return unsupportedTypeCloner
	}
}

/*
boolCloner 克隆bool类型
bool ----> ClonerFrom
bool ----> bool(true, false)
bool ----> any(true, false)
bool ----> string("true", "false")
bool ----> int(true:1, false:0)
bool ----> uint(true:1, false:0)
bool ----> float(true:1, false:0)
bool ----> pointer
*/
func boolCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	b := srcVal.Bool()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(b)
	}
	switch tv.Kind() {
	case reflect.Bool:
		tv.SetBool(b)
		return nil
	case reflect.Interface:
		return setAny(e, fks, tgtVal, srcVal, opts, tv, b)
	case reflect.String:
		tv.SetString(strconv.FormatBool(b))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(fks, tv, int64(boolMap[b]))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setInt2Uint(fks, tv, int64(boolMap[b]))
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, float64(boolMap[b]))
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, boolCloner)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
intCloner 克隆int类型
int ----> ClonerFrom
int ----> int(i)
int ----> any(i)
int ----> uint(i)
int ----> float(i)
int ----> bool(0->false, !0->true)
int ----> string("i")
int ----> pointer
*/
func intCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	i := srcVal.Int()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(i)
	}
	switch tv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(fks, tv, i)
	case reflect.Interface:
		return setAny(e, fks, tgtVal, srcVal, opts, tv, i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setInt2Uint(fks, tv, i)
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, float64(i))
	case reflect.Bool:
		tv.SetBool(i != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatInt(i, 10))
		return nil
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, intCloner)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
uintCloner 克隆uint类型
uint ----> ClonerFrom
uint ----> uint(u)
uint ----> any(u)
uint ----> int(u)
uint ----> float(u)
uint ----> bool(0->false, !0->true)
uint ----> string("u")
uint ----> pointer
*/
func uintCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	u := srcVal.Uint()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(u)
	}
	switch tv.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setUint(fks, tv, u)
	case reflect.Interface:
		return setAny(e, fks, tgtVal, srcVal, opts, tv, u)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setUint2Int(fks, tv, u)
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, float64(u))
	case reflect.Bool:
		tv.SetBool(u != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatUint(u, 10))
		return nil
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, uintCloner)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
floatCloner 克隆float类型
float ----> ClonerFrom
float ----> float(f)
float ----> any(f)
float ----> int(f)
float ----> uint(f)
float ----> bool(0->false, !0->true)
float ----> string("f")
float ----> pointer
*/
func floatCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	f := srcVal.Float()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(f)
	}
	switch tv.Kind() {
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, f)
	case reflect.Interface:
		return setAny(e, fks, tgtVal, srcVal, opts, tv, f)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setFloat2Int(fks, tv, f)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setFloat2Uint(fks, tv, f)
	case reflect.Bool:
		tv.SetBool(f != 0)
		return nil
	case reflect.String:
		tv.SetString(strconv.FormatFloat(f, 'f', -1, 64))
		return nil
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, floatCloner)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
stringCloner 克隆float类型
string ----> ClonerFrom
string ----> string("s")
string ----> []byte("s")
string ----> any("s")
string ----> bool("1", "t", "T", "true", "TRUE", "True"->false, "0", "f", "F", "false", "FALSE", "False"->true)
string ----> int("s")
string ----> uint("s")
string ----> float("s")
string ----> pointer
*/
func stringCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	s := srcVal.String()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(s)
	}
	switch tv.Kind() {
	case reflect.String:
		tv.SetString(s)
		return nil
	case reflect.Slice:
		if tv.Type().Elem().Kind() == reflect.Uint8 {
			tv.SetBytes([]byte(s))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Interface:
		return setAny(e, fks, tgtVal, srcVal, opts, tv, s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return newStringParseError(fks, tv.Type(), s, err)
		}
		tv.SetBool(b)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tv.Type(), s, err)
		}
		return setInt(fks, tv, i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return newStringParseError(fks, tv.Type(), s, err)
		}
		return setUint(fks, tv, u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newStringParseError(fks, tv.Type(), s, err)
		}
		return setFloat(fks, tv, f)
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, stringCloner)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
timeCloner 克隆time.Time类型
time.Time ----> ClonerFrom
time.Time ----> struct(time.Time)
time.Time ----> any(time.Time)
time.Time ----> string(time.RFC3339)
time.Time ----> []byte(time.RFC3339)
time.Time ----> int(time.Unix)
time.Time ----> uint(time.Unix)
time.Time ----> float(time.Unix)
time.Time ----> pointer
*/
func timeCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	t := srcVal.Interface().(time.Time)
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(t)
	}
	switch tv.Kind() {
	case reflect.Struct:
		if tv.Type() == timeType {
			tv.Set(reflect.ValueOf(t))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Interface:
		return setAny(e, fks, tgtVal, srcVal, opts, tv, t)
	case reflect.String:
		tv.SetString(t.Format(time.RFC3339))
		return nil
	case reflect.Slice:
		if tv.Type().Elem().Kind() == reflect.Uint8 {
			tv.SetBytes([]byte(t.Format(time.RFC3339)))
			return nil
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(fks, tv, opts.UnixTime(t))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return setUint(fks, tv, uint64(opts.UnixTime(t)))
	case reflect.Float32, reflect.Float64:
		return setFloat(fks, tv, float64(opts.UnixTime(t)))
	case reflect.Pointer:
		return setPointer(e, fks, tgtVal, srcVal, opts, tv, timeCloner)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

/*
structCloner 克隆 struct 类型
sql.NullBool ----> boolCloner
sql.NullByte ----> uintCloner
sql.NullInt16 ----> intCloner
sql.NullInt32 ----> intCloner
sql.NullInt64 ----> intCloner
sql.NullFloat64 ----> floatCloner
sql.NullString ----> stringCloner
sql.NullTime ----> timeCloner
struct ----> ClonerFrom
struct ----> struct
struct ----> any(map[string]any)
struct ----> map
struct ----> pointer
*/
func structCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}

	switch srcVal.Type() {
	case sqlNullBoolType:
		return boolCloner(e, fks, tgtVal, srcVal.FieldByName("Bool"), opts)
	case sqlNullByteType:
		return uintCloner(e, fks, tgtVal, srcVal.FieldByName("Byte"), opts)
	case sqlNullInt16Type:
		return intCloner(e, fks, tgtVal, srcVal.FieldByName("Int16"), opts)
	case sqlNullInt32Type:
		return intCloner(e, fks, tgtVal, srcVal.FieldByName("Int32"), opts)
	case sqlNullInt64Type:
		return intCloner(e, fks, tgtVal, srcVal.FieldByName("Int64"), opts)
	case sqlNullFloat64Type:
		return floatCloner(e, fks, tgtVal, srcVal.FieldByName("Float64"), opts)
	case sqlNullStringType:
		return stringCloner(e, fks, tgtVal, srcVal.FieldByName("String"), opts)
	case sqlNullTimeType:
		return timeCloner(e, fks, tgtVal, srcVal.FieldByName("Time"), opts)
	default:
		cloner, tv := indirectValue(tgtVal)
		if cloner != nil {
			return cloner.CloneFrom(srcVal.Interface())
		}
		switch tv.Kind() {
		case reflect.Struct:
			return struct2StructCloner(e, fks, tv, srcVal, opts)
		case reflect.Interface:
			if tv.NumMethod() == 0 {
				return struct2AnyCloner(e, fks, tv, srcVal, opts)
			}
			return hookCloner(e, fks, tgtVal, srcVal, opts)
		case reflect.Map:
			t := tv.Type()
			switch t.Key().Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
				reflect.Float32, reflect.Float64,
				reflect.String, reflect.Bool:
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
		case reflect.Pointer:
			return setPointer(e, fks, tgtVal, srcVal, opts, tv, structCloner)
		default:
			return hookCloner(e, fks, tgtVal, srcVal, opts)
		}
	}
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

			// 克隆src字段到tgt字段
			srcRecessiveField := srcFields.recessives[srcFields.recessivesFullNameIndex[fullName]]
			if err := srcRecessiveField.clonerFunc(e, append(slices.Clone(fks), srcKey), tgtRecessiveFieldVal, srcRecessiveFieldVal, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

func struct2AnyCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	m := make(map[string]any)
	srcType := srcVal.Type()
	srcFields := cachedTypeFields(srcType, opts, opts.SourceTagKey)
	for _, selfField := range srcFields.selfFields {
		// 查找src字段值
		srcDominantFieldVal, ok := findValue(srcVal, selfField)
		if !ok {
			continue
		}

		var vVal reflect.Value
		switch srcDominantFieldVal.Kind() {
		case reflect.String:
			vVal = reflect.ValueOf(new(string))
		case reflect.Bool:
			vVal = reflect.ValueOf(new(bool))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vVal = reflect.ValueOf(new(int64))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			vVal = reflect.ValueOf(new(uint64))
		case reflect.Float32, reflect.Float64:
			vVal = reflect.ValueOf(new(float64))
		case reflect.Slice, reflect.Array:
			vVal = reflect.ValueOf(new([]any))
		case reflect.Map, reflect.Struct:
			vVal = reflect.ValueOf(new(map[string]any))
		case reflect.Interface, reflect.Pointer:
			vVal = reflect.New(srcType)
		default:
			return errors.New("prototype: Unexpected type")
		}

		// 克隆src字段到tgt字段
		if err := selfField.clonerFunc(e, append(slices.Clone(fks), selfField.name), vVal, srcDominantFieldVal, opts); err != nil {
			return err
		}

		m[selfField.name] = vVal.Elem().Interface()
	}

	tgtVal.Set(reflect.ValueOf(m))
	return nil
}

func struct2MapCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	srcType := srcVal.Type()
	srcFields := cachedTypeFields(srcType, opts, opts.SourceTagKey)
	tgtType := tgtVal.Type()

	// 复制字段, 循环src字段
	var mapElem reflect.Value
	elemType := tgtType.Elem()
	for _, selfField := range srcFields.selfFields {
		fieldFks := append(slices.Clone(fks), selfField.name)

		// 查找src字段值
		srcDominantFieldVal, ok := findValue(srcVal, selfField)
		if !ok {
			continue
		}

		if !mapElem.IsValid() {
			mapElem = reflect.New(elemType).Elem()
		} else {
			mapElem.Set(reflect.Zero(elemType))
		}
		vVal := mapElem

		// 克隆src字段到tgt字段
		if err := selfField.clonerFunc(e, fieldFks, vVal, srcDominantFieldVal, opts); err != nil {
			return err
		}

		kType := tgtType.Key()
		var kVal reflect.Value
		switch {
		case reflect.PointerTo(kType).Implements(textUnmarshalerType):
			kVal = reflect.New(kType)
			if u, ok := kVal.Interface().(encoding.TextUnmarshaler); ok {
				err := u.UnmarshalText([]byte(selfField.name))
				if err != nil {
					return err
				}
			}
			kVal = kVal.Elem()
		case kType.Kind() == reflect.String:
			kVal = reflect.ValueOf(selfField.name).Convert(kType)
		case kType.Kind() == reflect.Bool:
			b, err := strconv.ParseBool(selfField.name)
			if err != nil {
				return err
			}
			kVal = reflect.ValueOf(b).Convert(kType)
		case slices.Contains(intKinds, kType.Kind()):
			i, err := strconv.ParseInt(selfField.name, 10, 64)
			if err != nil {
				return err
			}
			kVal = reflect.Zero(kType)
			if err := setInt(fks, kVal, i); err != nil {
				return err
			}
		case slices.Contains(uintKinds, kType.Kind()):
			u, err := strconv.ParseUint(selfField.name, 10, 64)
			if err != nil {
				return err
			}
			kVal = reflect.Zero(kType)
			if err := setUint(fks, kVal, u); err != nil {
				return err
			}
		case slices.Contains(floatKinds, kType.Kind()):
			f, err := strconv.ParseFloat(selfField.name, 64)
			if err != nil {
				return err
			}
			kVal = reflect.Zero(kType)
			if err := setFloat(fieldFks, kVal, f); err != nil {
				return err
			}
		default:
			return errors.New("prototype: Unexpected key type")
		}
		if !kVal.IsValid() {
			continue
		}
		tgtVal.SetMapIndex(kVal, vVal)
	}
	return nil
}

type kvPair struct {
	kVal   reflect.Value
	vVal   reflect.Value
	keyStr string
}

func (p *kvPair) resolve() error {
	if tm, ok := p.kVal.Interface().(encoding.TextMarshaler); ok {
		if p.kVal.Kind() == reflect.Pointer && p.kVal.IsNil() {
			return nil
		}
		buf, err := tm.MarshalText()
		p.keyStr = string(buf)
		return err
	}
	if str, ok := p.kVal.Interface().(fmt.Stringer); ok {
		if p.kVal.Kind() == reflect.Pointer && p.kVal.IsNil() {
			return nil
		}
		p.keyStr = str.String()
		return nil
	}
	switch p.kVal.Kind() {
	case reflect.String:
		p.keyStr = p.kVal.String()
		return nil
	case reflect.Bool:
		p.keyStr = strconv.FormatBool(p.kVal.Bool())
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p.keyStr = strconv.FormatInt(p.kVal.Int(), 10)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p.keyStr = strconv.FormatUint(p.kVal.Uint(), 10)
		return nil
	case reflect.Float32, reflect.Float64:
		p.keyStr = strconv.FormatFloat(p.kVal.Float(), 'f', -1, 64)
		return nil
	default:
		return errors.New("unexpected map key type")
	}
}

func mapCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if srcVal.IsNil() {
		return nil
	}

	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(srcVal.Interface())
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
	defer e.back()

	switch tv.Kind() {
	case reflect.Struct:
		return map2StructCloner(e, fks, tv, srcVal, opts)
	case reflect.Interface:
		if tv.NumMethod() == 0 {
			return map2AnyCloner(e, fks, tv, srcVal, opts)
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Map:
		if tv.IsNil() {
			tv.Set(reflect.MakeMap(tv.Type()))
		}
		return map2MapCloner(e, fks, tv, srcVal, opts)
	case reflect.Pointer:
		if tv.IsNil() {
			tv.Set(reflect.New(tv.Type().Elem()))
			tv = tv.Elem()
			return mapCloner(e, fks, tv, srcVal, opts)
		}
		return hookCloner(e, fks, tv, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func kvPairs(srcVal reflect.Value) ([]kvPair, error) {
	// Extract and sort the keys.
	kayValPairs := make([]kvPair, srcVal.Len())
	mapIter := srcVal.MapRange()
	for i := 0; mapIter.Next(); i++ {
		kayValPairs[i].kVal = mapIter.Key()
		kayValPairs[i].vVal = mapIter.Value()
		if err := kayValPairs[i].resolve(); err != nil {
			return nil, fmt.Errorf("prototype: map resolve error for type %q: %q", srcVal.Type().String(), err.Error())
		}
	}
	sort.Slice(kayValPairs, func(i, j int) bool { return strings.Compare(kayValPairs[i].keyStr, kayValPairs[j].keyStr) < 0 })
	return kayValPairs, nil
}

func map2MapCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	pairs, err := kvPairs(srcVal)
	if err != nil {
		return err
	}
	tgtType := tgtVal.Type()
	elemType := tgtType.Elem()
	var mapElem reflect.Value
	for _, pair := range pairs {
		if !mapElem.IsValid() {
			mapElem = reflect.New(elemType).Elem()
		} else {
			mapElem.Set(reflect.Zero(elemType))
		}
		vVal := mapElem

		if err := clonerByType(pair.vVal.Type(), true, opts)(e, append(slices.Clone(fks), pair.keyStr), vVal, pair.vVal, opts); err != nil {
			return err
		}

		kVal := reflect.New(tgtType.Key())
		if err := clonerByType(pair.kVal.Type(), true, opts)(e, append(slices.Clone(fks), pair.keyStr), kVal, pair.kVal, opts); err != nil {
			return err
		}
		kVal = kVal.Elem()

		if !kVal.IsValid() {
			continue
		}
		tgtVal.SetMapIndex(kVal, vVal)
	}
	return nil
}

func map2AnyCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	m := make(map[string]any)
	pairs, err := kvPairs(srcVal)
	if err != nil {
		return err
	}
	for _, pair := range pairs {
		valueCloner := clonerByType(pair.vVal.Type(), true, opts)

		var vVal reflect.Value
		switch pair.vVal.Kind() {
		case reflect.String:
			vVal = reflect.ValueOf(new(string))
		case reflect.Bool:
			vVal = reflect.ValueOf(new(bool))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vVal = reflect.ValueOf(new(int64))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			vVal = reflect.ValueOf(new(uint64))
		case reflect.Float32, reflect.Float64:
			vVal = reflect.ValueOf(new(float64))
		case reflect.Slice, reflect.Array:
			vVal = reflect.ValueOf(new([]any))
		case reflect.Map, reflect.Struct:
			vVal = reflect.ValueOf(new(map[string]any))
		case reflect.Interface, reflect.Pointer:
			vVal = reflect.New(pair.vVal.Type())
		default:
			return errors.New("prototype: Unexpected type")
		}

		if err := valueCloner(e, append(slices.Clone(fks), pair.keyStr), vVal, pair.vVal, opts); err != nil {
			return err
		}

		m[pair.keyStr] = vVal.Elem().Interface()
	}
	tgtVal.Set(reflect.ValueOf(m))
	return nil
}

func map2StructCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	pairs, err := kvPairs(srcVal)
	if err != nil {
		return err
	}
	tgtType := tgtVal.Type()
	tgtFields := cachedTypeFields(tgtType, opts, opts.TargetTagKey)

	for _, pair := range pairs {
		tgtField, ok := findField(tgtFields.selfNameIndex, tgtFields.selfFields, opts, pair.keyStr)
		if !ok {
			continue
		}
		tVal := tgtVal.FieldByIndex(tgtField.index)
		valueCloner := clonerByType(pair.vVal.Type(), true, opts)
		if err := valueCloner(e, append(slices.Clone(fks), pair.keyStr), tVal, pair.vVal, opts); err != nil {
			return err
		}
	}
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
	srcType := srcVal.Type()
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(srcVal.Interface())
	}
	if tv.Kind() == reflect.String && srcType.Elem().Kind() == reflect.Uint8 {
		builder := strings.Builder{}
		_, _ = builder.Write(srcVal.Bytes())
		tv.Set(reflect.ValueOf(builder.String()))
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
		}{ptr: srcVal.UnsafePointer(), len: srcVal.Len()}
		if e.isSeen(ptr) {
			return &UnsupportedValueError{Value: srcVal, Str: fmt.Sprintf("encountered a cycle via %s", srcVal.Type())}
		}
		e.remember(ptr)
		defer e.forget(ptr)
	}
	defer e.back()
	return arrayCloner(e, fks, tv, srcVal, opts)
}

func arrayCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	if !tgtVal.IsValid() {
		return nil
	}
	cloner, tv := indirectValue(tgtVal)
	if cloner != nil {
		return cloner.CloneFrom(srcVal.Interface())
	}
	switch tv.Kind() {
	case reflect.Array, reflect.Slice:
		srcLen := srcVal.Len()
		if tv.Kind() == reflect.Slice {
			tv.Set(reflect.MakeSlice(tv.Type(), srcLen, srcLen))
			tv.SetLen(0)
		}
		elemCloner := clonerByType(srcVal.Type().Elem(), true, opts)
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
			if err := elemCloner(e, append(slices.Clone(fks), strconv.FormatInt(int64(i), 10)), tgtItem, srcItem, opts); err != nil {
				return err
			}
			tv.Index(i).Set(tgtItem)
		}
		return nil
	case reflect.Interface:
		if tv.NumMethod() == 0 {
			return array2AnyCloner(e, fks, tv, srcVal, opts)
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	case reflect.Pointer:
		if tv.IsNil() {
			tv.Set(reflect.New(tv.Type().Elem()))
			tv = tv.Elem()
			return arrayCloner(e, fks, tv, srcVal, opts)
		}
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	default:
		return hookCloner(e, fks, tgtVal, srcVal, opts)
	}
}

func array2AnyCloner(e *cloneContext, fks []string, tgtVal, srcVal reflect.Value, opts *options) error {
	elemEnc := clonerByType(tgtVal.Type().Elem(), true, opts)
	srcLen := srcVal.Len()
	// 创建一个切片
	tgtSlice := make([]any, 0, srcLen)
	// 将src的元素逐个拷贝到tgt
	for i := 0; i < srcLen; i++ {
		var tgtItem any
		tgtItemVal := reflect.ValueOf(&tgtItem)
		srcItemVal := srcVal.Index(i)
		if err := elemEnc(e, append(slices.Clone(fks), convx.ToString(i)), tgtItemVal, srcItemVal, opts); err != nil {
			return err
		}
		tgtSlice = append(tgtSlice, tgtItem)
	}
	// 设置tgtVal
	tgtVal.Set(reflect.ValueOf(tgtSlice))
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
	return clonerByValue(srcVal, opts)(e, fks, tgtVal, srcVal, opts)
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
	defer e.back()
	cloner := clonerByType(srcVal.Type().Elem(), true, opts)
	if err := cloner(e, fks, tgtVal, srcVal.Elem(), opts); err != nil {
		return err
	}
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
	return &UnsupportedTypeError{SourceType: srcVal.Type(), TargetType: tgtVal.Type()}
}
