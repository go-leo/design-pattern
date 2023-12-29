package prototype

import (
	"github.com/go-leo/design-pattern/prototype/internal"
	"golang.org/x/exp/slices"
	"reflect"
	"strings"
	"sync"
)

type structInfo struct {
	Type reflect.Type
	// 自有字段索引 label ---> 字段
	FieldIndexes fieldInfoIndexes
	// 方法 name ---> 字段
	StructMethodIndexes  map[string]*methodInfo
	PointerMethodIndexes map[string]*methodInfo
	// 所有字段索引 label ---> 字段
	AllFieldIndexes fieldInfoIndexes
}

func (s *structInfo) Analysis(opts *options) *structInfo {
	// 自有字段分析
	s.AnalysisFields(opts)
	// 方法分析
	s.AnalysisMethods(opts)
	// 所有（自有和嵌入的）字段分析
	s.AnalysisAllFields(opts)
	// 排序字段
	s.SortFields(opts)
	return s
}

func (s *structInfo) AnalysisFields(opts *options) {
	// 字段分析
	for i := 0; i < s.Type.NumField(); i++ {
		field := &fieldInfo{StructField: s.Type.Field(i)}
		field.Analysis(opts)
		// 忽略字段
		if field.IsIgnore {
			continue
		}
		s.FieldIndexes[field.Label] = append(s.FieldIndexes[field.Label], field)
	}
}

func (s *structInfo) AnalysisMethods(opts *options) {
	// 方法分析
	typ := s.Type
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		// 忽略未导出方法
		if !method.IsExported() {
			continue
		}
		lowerMethodName := strings.ToLower(method.Name)
		lowerGetterPrefix := strings.ToLower(opts.GetterPrefix)
		lowerSetterPrefix := strings.ToLower(opts.SetterPrefix)
		if strings.HasPrefix(lowerMethodName, lowerGetterPrefix) ||
			strings.HasSuffix(lowerMethodName, lowerSetterPrefix) {
			s.StructMethodIndexes[method.Name] = &methodInfo{Method: method}
		}
	}
	ptrType := reflect.PointerTo(typ)
	for i := 0; i < ptrType.NumMethod(); i++ {
		method := ptrType.Method(i)
		// 忽略未导出方法
		if !method.IsExported() {
			continue
		}
		lowerMethodName := strings.ToLower(method.Name)
		lowerGetterPrefix := strings.ToLower(opts.GetterPrefix)
		lowerSetterPrefix := strings.ToLower(opts.SetterPrefix)
		if strings.HasPrefix(lowerMethodName, lowerGetterPrefix) ||
			strings.HasSuffix(lowerMethodName, lowerSetterPrefix) {
			s.PointerMethodIndexes[method.Name] = &methodInfo{Method: method}
		}
	}
}

func (s *structInfo) AnalysisAllFields(opts *options) {
	for _, fields := range s.FieldIndexes {
		for _, field := range fields {
			// 非匿名字段， 直接加入 AllFieldIndexes
			if !field.Anonymous {
				s.AllFieldIndexes[field.Label] = append(s.AllFieldIndexes[field.Label], field)
				continue
			}

			// 匿名字段
			var anmStruct *structInfo
			if field.Type.Kind() == reflect.Struct {
				anmStruct = cachedStruct(field.Type, opts)
			}
			if field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct {
				anmStruct = cachedStruct(field.Type.Elem(), opts)
			}

			// 匿名、非结构体和结构体指针字段，直接加入 AllFieldIndexes
			if anmStruct == nil {
				s.AllFieldIndexes[field.Label] = append(s.AllFieldIndexes[field.Label], field)
				continue
			}

			// 匿名、结构体和结构体指针字段的AllFieldIndexes 加入 AllFieldIndexes
			for _, anmFields := range anmStruct.AllFieldIndexes {
				for _, anmField := range anmFields {
					s.AllFieldIndexes[anmField.Label] = append(s.AllFieldIndexes[anmField.Label], anmField.Clone().Unshift(field))
				}
			}
		}

	}
}

func (s *structInfo) SortFields(opts *options) {
	for _, fields := range s.FieldIndexes {
		fields.Sort(opts)
	}
	for _, fields := range s.AllFieldIndexes {
		fields.Sort(opts)
	}
}

func (s *structInfo) FindField(label string, field *fieldInfo, opts *options) (*fieldInfo, bool) {
	// 精准匹配
	if field, ok := s.ExactMatchField(label, field, s.AllFieldIndexes, opts); ok {
		return field, true
	}
	if field, ok := s.ExactMatchField(label, field, s.FieldIndexes, opts); ok {
		return field, true
	}

	// 模糊匹配
	if field, ok := s.CaseFoldMatchField(label, field, s.AllFieldIndexes, opts); ok {
		return field, true
	}
	if field, ok := s.CaseFoldMatchField(label, field, s.FieldIndexes, opts); ok {
		return field, true
	}
	return nil, false
}

func (s *structInfo) ExactMatchField(label string, field *fieldInfo, indexes fieldInfoIndexes, _ *options) (*fieldInfo, bool) {
	if fields, ok := indexes[label]; ok {
		if field == nil {
			return fields[0], true
		}
		return s.MatchField(fields, field, func(a, b string) bool { return a == b })
	}
	return nil, false
}

func (s *structInfo) CaseFoldMatchField(label string, field *fieldInfo, indexes fieldInfoIndexes, opts *options) (*fieldInfo, bool) {
	for fieldLabel, fields := range indexes {
		if !opts.EqualFold(label, fieldLabel) {
			continue
		}
		return s.MatchField(fields, field, opts.EqualFold)
	}
	return nil, false
}

func (*structInfo) MatchField(fields fieldInfos, field *fieldInfo, equals func(a, b string) bool) (*fieldInfo, bool) {
	if field == nil {
		return fields[0], true
	}
	var matchField *fieldInfo
	for _, f := range fields {
		if !equals(f.Labels, field.Labels) {
			continue
		}
		if matchField == nil {
			matchField = f
		}
		if !equals(f.Names, field.Names) {
			continue
		}
		return field, true
	}
	if matchField != nil {
		return matchField, true
	}
	return fields[0], true
}

func (s *structInfo) FindMethod(label string, opts *options, methodIndexes map[string]*methodInfo) (*methodInfo, bool) {
	// 完全匹配
	if m, ok := methodIndexes[label]; ok {
		return m, true
	}
	// 模糊匹配
	for name, m := range methodIndexes {
		if opts.EqualFold(name, label) {
			return m, true
		}
	}
	return nil, false
}

func (s *structInfo) FindGetter(label string, v reflect.Value, opts *options) (*methodInfo, reflect.Value, bool) {
	label = opts.GetterPrefix + label
	var ok bool
	var method *methodInfo
	if v.CanAddr() {
		method, ok = s.FindMethod(label, opts, s.PointerMethodIndexes)
		v = v.Addr()
	} else {
		method, ok = s.FindMethod(label, opts, s.StructMethodIndexes)
	}
	if !ok {
		return nil, reflect.Value{}, false
	}
	methodVal := v.Method(method.Index)
	ok = method.CheckGetter(methodVal)
	return method, methodVal, ok
}

func (s *structInfo) FindSetter(label string, v reflect.Value, opts *options) (*methodInfo, reflect.Value, bool) {
	label = opts.SetterPrefix + label
	var method *methodInfo
	var ok bool
	if v.CanAddr() {
		v = v.Addr()
		method, ok = s.FindMethod(label, opts, s.PointerMethodIndexes)
	} else {
		method, ok = s.FindMethod(label, opts, s.StructMethodIndexes)
	}
	if !ok {
		return nil, reflect.Value{}, false
	}
	methodVal := v.Method(method.Index)
	ok = method.CheckSetter(methodVal)
	return method, methodVal, ok
}

func (s *structInfo) RangeFields(f func(label string, field *fieldInfo) error) error {
	for label, fields := range s.FieldIndexes {
		for _, field := range fields {
			if err := f(label, field); err != nil {
				return err
			}
		}
	}
	return nil
}

type fieldInfo struct {
	reflect.StructField
	Indexes  []int
	Names    string
	Labels   string
	WithTag  bool
	Label    string
	Options  []string
	IsIgnore bool
}

func (f *fieldInfo) Analysis(opts *options) *fieldInfo {
	f.Indexes = slices.Clone(f.StructField.Index)
	f.Names = f.StructField.Name
	tagValue := f.Tag.Get(opts.TagKey)
	// 如果是tag是"-",则忽略该字段
	if tagValue == "-" {
		f.Labels = ""
		f.WithTag = false
		f.Label = ""
		f.Options = []string{}
		f.IsIgnore = true
		return f
	}
	// 没找到tag，或者value为空，默认的字段名
	if len(tagValue) <= 0 {
		f.Labels = f.StructField.Name
		f.WithTag = false
		f.Label = f.StructField.Name
		f.Options = []string{}
		f.IsIgnore = false
		return f
	}
	// 以","分割value，
	values := strings.Split(tagValue, ",")
	f.Labels = values[0]
	f.WithTag = true
	f.Label = values[0]
	f.Options = slices.Clone(values[1:])
	f.IsIgnore = false
	return f
}

func (f *fieldInfo) Clone() *fieldInfo {
	cloned := fieldInfo{
		StructField: f.StructField,
		Indexes:     slices.Clone(f.Indexes),
		Names:       f.Names,
		Labels:      f.Labels,
		WithTag:     f.WithTag,
		Label:       f.Label,
		Options:     f.Options,
		IsIgnore:    f.IsIgnore,
	}
	return &cloned
}

func (f *fieldInfo) Unshift(parent *fieldInfo) *fieldInfo {
	f.Indexes = append(slices.Clone(parent.Indexes), f.Indexes...)
	f.Names = parent.Names + "." + f.Names
	f.Labels = parent.Labels + "." + f.Labels
	return f
}

func (f *fieldInfo) FindGettableValue(val reflect.Value) (reflect.Value, bool) {
	outVal := val
	for _, i := range f.Indexes {
		outVal = outVal.Field(i)
		if outVal.Kind() == reflect.Pointer {
			if outVal.IsNil() {
				return reflect.Value{}, false
			}
			outVal = outVal.Elem()
		}
	}
	return outVal, true
}

func (f *fieldInfo) FindSettableValue(val reflect.Value) (reflect.Value, bool) {
	outVal := val
	for _, i := range f.Indexes {
		outVal = outVal.Field(i)
		if outVal.Kind() == reflect.Pointer {
			if outVal.IsNil() {
				if !outVal.CanSet() {
					return outVal, false
				}
				outVal.Set(reflect.New(outVal.Type().Elem()))
			}
			outVal = outVal.Elem()
		}
	}
	return outVal, true
}

func (f *fieldInfo) ContainsOption(option string) bool {
	return slices.Contains(f.Options, option)
}

type fieldInfos []*fieldInfo

func (s fieldInfos) Sort(opts *options) {
	slices.SortFunc(s, func(a, b *fieldInfo) bool {
		// 深度越浅越优先
		if len(a.Indexes) != len(b.Indexes) {
			return len(a.Indexes) < len(b.Indexes)
		}
		// 打上标签的优先
		if a.WithTag != b.WithTag {
			return a.WithTag
		}
		// 标签名与字段名相同的优先
		if a.Name != b.Name {
			if opts.EqualFold(a.Name, a.Label) {
				return true
			}
			if opts.EqualFold(b.Name, b.Label) {
				return true
			}
		}
		// 字段index越小越优先
		for i, aIndex := range a.Indexes {
			bIndex := b.Indexes[i]
			if aIndex != bIndex {
				return aIndex < bIndex
			}
		}
		// 字段类型越基础有优先
		return internal.KindOrder[a.Type.Kind()] < internal.KindOrder[b.Type.Kind()]
	})
}

func (s fieldInfos) Merge(infos fieldInfos) fieldInfos {
	r := s
OUT:
	for _, info := range infos {
		for _, field := range s {
			if info.Label == field.Label && info.Labels == field.Labels {
				continue OUT
			}
		}
		r = append(r, info)
	}
	return r
}

type fieldInfoIndexes map[string]fieldInfos

func mergeFieldInfoIndexes(opts *options, allIndexes ...fieldInfoIndexes) fieldInfoIndexes {
	r := make(fieldInfoIndexes)
	for _, indexes := range allIndexes {
		for label, fields := range indexes {
			r[label] = r[label].Merge(fields)
		}
	}
	for _, fields := range r {
		fields.Sort(opts)
	}
	return r
}

type methodInfo struct {
	reflect.Method
}

// CheckGetter
// func(x *Obj)Method() string
// func(x *Obj)Method() (string, error)
func (m *methodInfo) CheckGetter(method reflect.Value) bool {
	if !method.IsValid() {
		return false
	}
	methodType := method.Type()
	if methodType.NumIn() > 0 {
		return false
	}
	if methodType.NumOut() == 1 {
		return true
	}
	if methodType.NumOut() == 2 && methodType.Out(1) == errorType {
		return true
	}
	return false
}

// CheckSetter
// func(x *Obj)Method(string)
// func(x *Obj)Method(string) error
func (m *methodInfo) CheckSetter(method reflect.Value) bool {
	if !method.IsValid() {
		return false
	}
	methodType := method.Type()
	if methodType.NumIn() != 1 {
		return false
	}
	if methodType.NumOut() == 0 {
		return true
	}
	if methodType.NumOut() == 1 && methodType.Out(0) == errorType {
		return true
	}
	return false
}

func (m *methodInfo) InvokeGetter(getter reflect.Value) (reflect.Value, error) {
	outValues := getter.Call([]reflect.Value{})
	if len(outValues) == 1 {
		return outValues[0], nil

	}
	if err, ok := outValues[1].Interface().(error); ok && err != nil {
		return reflect.Value{}, err
	}
	return outValues[0], nil
}

func (m *methodInfo) InvokeSetter(inVal, setter reflect.Value) error {
	outValues := setter.Call([]reflect.Value{inVal})
	if len(outValues) <= 0 {
		return nil
	}
	if err, ok := outValues[0].Interface().(error); ok && err != nil {
		return err
	}
	return nil
}

func newStructInfo(t reflect.Type) *structInfo {
	return &structInfo{
		Type:                 t,
		FieldIndexes:         make(fieldInfoIndexes),
		StructMethodIndexes:  make(map[string]*methodInfo),
		PointerMethodIndexes: make(map[string]*methodInfo),
		AllFieldIndexes:      make(fieldInfoIndexes),
	}
}

var structCache sync.Map

func cachedStruct(t reflect.Type, opts *options) *structInfo {
	if f, ok := structCache.Load(t); ok {
		return f.(*structInfo)
	}
	f, _ := structCache.LoadOrStore(t, newStructInfo(t).Analysis(opts))
	return f.(*structInfo)
}

type mapEntry struct {
	KeyVal  reflect.Value
	Label   string
	ValVal  reflect.Value
	ValType reflect.Type
}

type mapEntries []mapEntry

func (s mapEntries) Sort() {
	slices.SortFunc(s, func(a, b mapEntry) bool {
		if a.ValType.Kind() != b.ValType.Kind() {
			if a.ValType.Kind() == reflect.Struct || a.ValType.Kind() == reflect.Map {
				return false
			}
		}
		return strings.Compare(a.Label, b.Label) < 0
	})
}

func newMapEntries(srcMapIter *reflect.MapIter) (mapEntries, error) {
	entries := make(mapEntries, 0)
	for srcMapIter.Next() {
		valVal := srcMapIter.Value()
		if !valVal.IsValid() || valVal.IsNil() {
			continue
		}
		keyVal := srcMapIter.Key()
		label, err := stringify(keyVal)
		if err != nil {
			return nil, newStringifyError(keyVal.Type(), err)
		}
		entries = append(entries, mapEntry{
			KeyVal:  keyVal,
			Label:   label,
			ValVal:  valVal,
			ValType: indirectValue(valVal).Type(),
		})
	}
	entries.Sort()
	return entries, nil
}
