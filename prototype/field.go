package prototype

import (
	"fmt"
	"github.com/go-leo/design-pattern/prototype/internal"
	"golang.org/x/exp/slices"
	"reflect"
	"strings"
	"sync"
)

type _StructInfo struct {
	Type reflect.Type
	// 自有直接的字段
	FieldIndexes _FieldIndexes
	// 非嵌入结构体字段的所有字段
	AllFieldIndexes _FieldIndexes
	// 嵌入结构体字段的所有字段
	EmbedFieldIndexes _FieldIndexes
	BaseMethods       _MethodInfos
	PointerMethods    _MethodInfos
}

func (s *_StructInfo) Analysis(opts *options) *_StructInfo {
	s.AnalysisMethods(opts)
	s.AnalysisFields(opts)
	s.AnalysisEmbeddedFields(opts)
	return s
}

func (s *_StructInfo) AnalysisMethods(opts *options) {
	// 方法分析
	typ := s.Type
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		// 忽略未导出方法
		if !method.IsExported() {
			continue
		}
		s.BaseMethods = append(s.BaseMethods, &_MethodInfo{Method: method, ReceiverType: typ})
	}
	ptrType := reflect.PointerTo(typ)
	for i := 0; i < ptrType.NumMethod(); i++ {
		method := ptrType.Method(i)
		// 忽略未导出方法
		if !method.IsExported() {
			continue
		}
		s.PointerMethods = append(s.PointerMethods, &_MethodInfo{Method: method, ReceiverType: typ})
	}
}

func (s *_StructInfo) AnalysisFields(opts *options) {
	// 字段分析
	for i := 0; i < s.Type.NumField(); i++ {
		field := &_FieldInfo{
			Parent:      s,
			StructField: s.Type.Field(i),
		}
		// 忽略字段
		if field.Analysis(opts).Ignored {
			continue
		}
		s.FieldIndexes[field.Label] = append(s.FieldIndexes[field.Label], field)
	}
}

func (s *_StructInfo) AnalysisEmbeddedFields(opts *options) {
	for _, infos := range s.FieldIndexes {
		for _, field := range infos {
			// 非匿名
			if !field.Anonymous {
				s.AllFieldIndexes[field.Label] = append(s.AllFieldIndexes[field.Label], field)
				continue
			}
			// 匿名字段

			var nestedStruct *_StructInfo
			if field.Type.Kind() == reflect.Struct {
				nestedStruct = _CachedStructInfo(field.Type, opts)
			} else if field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct {
				nestedStruct = _CachedStructInfo(field.Type.Elem(), opts)
			} else {
				s.AllFieldIndexes[field.Label] = append(s.AllFieldIndexes[field.Label], field)
				continue
			}

			for _, anmFields := range nestedStruct.AllFieldIndexes {
				for _, anmField := range anmFields {
					s.AllFieldIndexes[anmField.Label] = append(s.AllFieldIndexes[anmField.Label], anmField.Clone().Unshift(field))
				}
			}

			s.EmbedFieldIndexes[field.Label] = append(s.EmbedFieldIndexes[field.Label], field)
			for _, embedFields := range nestedStruct.EmbedFieldIndexes {
				for _, embedField := range embedFields {
					s.EmbedFieldIndexes[field.Label] = append(s.EmbedFieldIndexes[field.Label], embedField.Clone().Unshift(field))
				}
			}

		}
	}
}

func (s *_StructInfo) RangeFields(f func(label string, field *_FieldInfo) error) error {
	for label, fields := range s.FieldIndexes {
		for _, field := range fields {
			if err := f(label, field); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *_StructInfo) RangeAllFields(f func(label string, field *_FieldInfo) error) error {
	for label, fields := range s.AllFieldIndexes {
		for _, field := range fields {
			if err := f(label, field); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *_StructInfo) RangeEmbedFields(f func(label string, field *_FieldInfo) error) error {
	for label, fields := range s.EmbedFieldIndexes {
		for _, field := range fields {
			if err := f(label, field); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *_StructInfo) FindValueByLabel(label string) *_FieldInfo {
	if fields, ok := s.AllFieldIndexes[label]; ok {
		return fields[0]
	}
	if fields, ok := s.EmbedFieldIndexes[label]; ok {
		return fields[0]
	}
	return nil
}

func (s *_StructInfo) FindGettableValue(structVal reflect.Value, field *_FieldInfo) (reflect.Value, reflect.Value, bool) {
	parentVal := structVal
	fieldVal := structVal
	for _, i := range field.Indexes {
		fieldVal = fieldVal.Field(i)
		if fieldVal.Kind() == reflect.Pointer {
			if fieldVal.IsNil() {
				return parentVal, fieldVal, false
			}
			fieldVal = fieldVal.Elem()
		}
	}
	return parentVal, fieldVal, true
}

func (s *_StructInfo) FindSettableValue(structVal reflect.Value, field *_FieldInfo) (reflect.Value, reflect.Value, bool) {
	parentVal := structVal
	fieldVal := structVal
	for _, i := range field.Indexes {
		parentVal = fieldVal
		fieldVal = fieldVal.Field(i)
		if fieldVal.Kind() == reflect.Pointer {
			if fieldVal.IsNil() {
				if !fieldVal.CanSet() {
					return parentVal, fieldVal, false
				}
				fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			}
			fieldVal = fieldVal.Elem()
		}
	}
	return parentVal, fieldVal, true
}

type _StructInfos []*_StructInfo

type _MethodInfo struct {
	reflect.Method
	ReceiverType reflect.Type
}

// IsGetter
// func(x *Obj)Method() string
// func(x *Obj)Method(context.Context) (string, error)
func (m *_MethodInfo) IsGetter() bool {
	methodType := m.Type
	if methodType.NumIn() == 1 && methodType.NumOut() == 1 {
		return true
	}
	if methodType.NumIn() == 2 && methodType.NumOut() == 2 &&
		methodType.In(1) == contextType && methodType.Out(1) == errorType {
		return true
	}
	return false
}

// IsSetter
// func(x *Obj)Method(string)
// func(x *Obj)Method(context.Context, string) error
func (m *_MethodInfo) IsSetter() bool {
	methodType := m.Type
	if methodType.NumIn() == 2 && methodType.NumOut() == 0 {
		return true
	}
	if methodType.NumIn() == 3 && methodType.NumOut() == 1 &&
		methodType.In(1) == contextType && methodType.Out(0) == errorType {
		return true
	}
	return false
}

func (m *_MethodInfo) InvokeGetter(getter reflect.Value, opts *options) (reflect.Value, error) {
	inValues := make([]reflect.Value, 0)
	if m.Type.NumIn() == 2 {
		inValues = append(inValues, reflect.ValueOf(opts.Context))
	}
	outValues := getter.Call(inValues)
	if len(outValues) == 1 {
		return outValues[0], nil

	}
	if err, ok := outValues[1].Interface().(error); ok && err != nil {
		return reflect.Value{}, err
	}
	return outValues[0], nil
}

func (m *_MethodInfo) InvokeSetter(inVal, setter reflect.Value, opts *options) error {
	inValues := make([]reflect.Value, 0)
	if m.Type.NumIn() == 3 {
		inValues = append(inValues, reflect.ValueOf(opts.Context))
	}
	inValues = append(inValues, inVal)
	outValues := setter.Call(inValues)
	if len(outValues) <= 0 {
		return nil
	}
	if err, ok := outValues[0].Interface().(error); ok && err != nil {
		return err
	}
	return nil
}

type _MethodInfos []*_MethodInfo

type _FieldInfo struct {
	reflect.StructField
	Parent        *_StructInfo
	Indexes       []int
	Names         []string
	WithTag       bool
	Ignored       bool
	Label         string
	Labels        []string
	Options       []string
	BaseGetter    *_MethodInfo
	BaseSetter    *_MethodInfo
	PointerGetter *_MethodInfo
	PointerSetter *_MethodInfo
}

func (f *_FieldInfo) Analysis(opts *options) *_FieldInfo {
	tagValue := f.Tag.Get(opts.TagKey)
	// 如果是tag是"-",则忽略该字段
	if tagValue == "-" {
		f.Ignored = true
		return f
	}
	f.Ignored = false
	f.Indexes = slices.Clone(f.StructField.Index)
	f.Names = []string{f.Name}

	for _, method := range f.Parent.BaseMethods {
		if strings.EqualFold(method.Name, opts.GetterPrefix+f.Name) && method.IsGetter() {
			f.BaseGetter = method
		} else if strings.EqualFold(method.Name, opts.SetterPrefix+f.Name) && method.IsSetter() {
			f.BaseSetter = method
		}
	}
	for _, method := range f.Parent.PointerMethods {
		if strings.EqualFold(method.Name, opts.GetterPrefix+f.Name) && method.IsGetter() {
			f.PointerGetter = method
		} else if strings.EqualFold(method.Name, opts.SetterPrefix+f.Name) && method.IsSetter() {
			f.PointerSetter = method
		}
	}

	// 没找到tag，或者value为空，默认的字段名
	if len(tagValue) <= 0 {
		f.WithTag = false
		f.Label = f.Name
		f.Labels = []string{f.Label}
		f.Options = []string{}
		return f
	}
	// 以","分割value，
	values := strings.Split(tagValue, ",")
	f.WithTag = true
	f.Label = values[0]
	f.Labels = []string{f.Label}
	f.Options = slices.Clone(values[1:])
	return f
}

func (f *_FieldInfo) Clone() *_FieldInfo {
	cloned := _FieldInfo{
		StructField:   f.StructField,
		Parent:        f.Parent,
		Indexes:       slices.Clone(f.Indexes),
		Names:         slices.Clone(f.Names),
		Labels:        slices.Clone(f.Labels),
		WithTag:       f.WithTag,
		Ignored:       f.Ignored,
		Label:         f.Label,
		Options:       f.Options,
		BaseGetter:    f.BaseGetter,
		BaseSetter:    f.BaseSetter,
		PointerGetter: f.PointerGetter,
		PointerSetter: f.PointerSetter,
	}
	return &cloned
}

func (f *_FieldInfo) Unshift(parent *_FieldInfo) *_FieldInfo {
	f.Indexes = append(slices.Clone(parent.Indexes), f.Indexes...)
	f.Names = append(slices.Clone(parent.Names), f.Names...)
	f.Labels = append(slices.Clone(parent.Labels), f.Labels...)
	return f
}

func (f *_FieldInfo) GetValue(structVal reflect.Value, s *_StructInfo, opts *options) (reflect.Value, error) {
	parentVal, fieldVal, ok := s.FindGettableValue(structVal, f)
	if ok {
		return reflect.Value{}, nil
	}
	var getter reflect.Value
	var methodInfo *_MethodInfo
	if f.PointerGetter != nil && parentVal.CanAddr() {
		parentVal = parentVal.Addr()
		methodInfo = f.PointerGetter
	}
	if f.BaseGetter != nil {
		methodInfo = f.BaseGetter
	}
	getter = parentVal.Method(methodInfo.Index)
	if methodInfo != nil && getter.IsValid() {
		// 方法克隆
		return methodInfo.InvokeGetter(getter, opts)
	}
	return fieldVal, nil
}

func (f *_FieldInfo) SetValue(s *_StructInfo, structVal reflect.Value, opts *options, setFunc func(in reflect.Value) error) error {
	parentVal, fieldVal, ok := s.FindSettableValue(structVal, f)
	if !ok {
		return nil
	}
	var setter reflect.Value
	var methodInfo *_MethodInfo
	if f.PointerSetter != nil && parentVal.CanAddr() {
		methodInfo = f.PointerSetter
		parentVal = parentVal.Addr()
	}
	if f.BaseSetter != nil {
		methodInfo = f.BaseSetter
	}
	if methodInfo == nil {
		return setFunc(fieldVal)
	}
	setter = parentVal.Method(methodInfo.Index)
	if !setter.IsValid() {
		return setFunc(fieldVal)
	}
	inVal := reflect.New(methodInfo.Type.In(methodInfo.Type.NumIn() - 1)).Elem()
	if err := setFunc(inVal); err != nil {
		return err
	}
	// 方法
	return methodInfo.InvokeSetter(inVal, setter, opts)
}

type _FieldInfos []*_FieldInfo

type _FieldIndexes map[string]_FieldInfos

func _NewStructInfo(typ reflect.Type) *_StructInfo {
	return &_StructInfo{
		Type:              typ,
		FieldIndexes:      make(_FieldIndexes),
		AllFieldIndexes:   make(_FieldIndexes),
		EmbedFieldIndexes: make(_FieldIndexes),
		BaseMethods:       make(_MethodInfos, 0),
		PointerMethods:    make(_MethodInfos, 0),
	}
}

func _CachedStructInfo(typ reflect.Type, opts *options) *_StructInfo {
	if f, ok := structCache.Load(typ); ok {
		return f.(*_StructInfo)
	}
	f, _ := structCache.LoadOrStore(typ, _NewStructInfo(typ).Analysis(opts))
	return f.(*_StructInfo)
}

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

type _MapEntry struct {
	KeyVal  reflect.Value
	Label   string
	ValVal  reflect.Value
	ValType reflect.Type
}

type _MapEntries []_MapEntry

func (s _MapEntries) Sort() {
	slices.SortFunc(s, func(a, b _MapEntry) bool {
		if a.ValType.Kind() != b.ValType.Kind() {
			return _KindOrder[a.ValType.Kind()] < _KindOrder[b.ValType.Kind()]
		}
		return strings.Compare(a.Label, b.Label) < 0
	})
}

func newMapEntries(srcMapIter *reflect.MapIter) (_MapEntries, error) {
	entries := make(_MapEntries, 0)
	for srcMapIter.Next() {
		keyVal := srcMapIter.Key()
		label, err := stringify(keyVal)
		if err != nil {
			return nil, newStringifyError(keyVal.Type(), err)
		}

		valVal := srcMapIter.Value()
		fmt.Println(valVal.Type())
		if !valVal.IsValid() {
			continue
		}
		if valVal.IsNil() {

		}
		entries = append(entries, _MapEntry{
			KeyVal:  keyVal,
			Label:   label,
			ValVal:  valVal,
			ValType: indirectValue(valVal).Type(),
		})
	}
	entries.Sort()
	return entries, nil
}
