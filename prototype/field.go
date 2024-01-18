package prototype

import (
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
	s.Sort(opts)
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
		s.BaseMethods = append(s.BaseMethods, &_MethodInfo{Method: method})
	}
	ptrType := reflect.PointerTo(typ)
	for i := 0; i < ptrType.NumMethod(); i++ {
		method := ptrType.Method(i)
		// 忽略未导出方法
		if !method.IsExported() {
			continue
		}
		s.PointerMethods = append(s.PointerMethods, &_MethodInfo{Method: method})
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

func (s *_StructInfo) Sort(opts *options) {
	s.FieldIndexes.Sort(opts)
	s.AllFieldIndexes.Sort(opts)
	s.EmbedFieldIndexes.Sort(opts)
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

func (s *_StructInfo) FindFieldByLabel(label string, opts *options) *_FieldInfo {
	if field := s.AllFieldIndexes.FindValueByLabel(label, opts); field != nil {
		return field
	}
	return s.EmbedFieldIndexes.FindValueByLabel(label, opts)
}

func (s *_StructInfo) FindFieldByField(field *_FieldInfo, opts *options) *_FieldInfo {
	if f := s.AllFieldIndexes.FindValueByField(field, opts); f != nil {
		return f
	}
	return s.EmbedFieldIndexes.FindValueByField(field, opts)
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
	exported := true
	for _, i := range field.Indexes {
		if exported {
			parentVal = fieldVal
		}
		fieldStruct := fieldVal.Type().Field(i)
		exported = fieldStruct.IsExported()
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

type _FieldInfo struct {
	reflect.StructField
	Parent    *_StructInfo
	Indexes   []int
	Names     []string
	Tagged    bool
	Ignored   bool
	Label     string
	Labels    []string
	OmitEmpty bool
	//Options       []string
	BaseGetter    string
	BaseSetter    string
	PointerGetter string
	PointerSetter string
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
			f.BaseGetter = method.Name
		} else if strings.EqualFold(method.Name, opts.SetterPrefix+f.Name) && method.IsSetter() {
			f.BaseSetter = method.Name
		}
	}
	for _, method := range f.Parent.PointerMethods {
		if strings.EqualFold(method.Name, opts.GetterPrefix+f.Name) && method.IsGetter() {
			f.PointerGetter = method.Name
		} else if strings.EqualFold(method.Name, opts.SetterPrefix+f.Name) && method.IsSetter() {
			f.PointerSetter = method.Name
		}
	}

	// 没找到tag，或者value为空，默认的字段名
	if len(tagValue) <= 0 {
		f.Tagged = false
		f.Label = f.Name
		f.Labels = []string{f.Label}
		f.OmitEmpty = false
		//f.Options = []string{}
		return f
	}
	// 以","分割value，
	values := strings.Split(tagValue, ",")
	f.Tagged = true
	f.Label = values[0]
	f.Labels = []string{f.Label}
	f.OmitEmpty = slices.Contains(values[1:], "omitempty")
	return f
}

func (f *_FieldInfo) Clone() *_FieldInfo {
	cloned := _FieldInfo{
		StructField:   f.StructField,
		Parent:        f.Parent,
		Indexes:       slices.Clone(f.Indexes),
		Names:         slices.Clone(f.Names),
		Tagged:        f.Tagged,
		Ignored:       f.Ignored,
		Label:         f.Label,
		Labels:        slices.Clone(f.Labels),
		OmitEmpty:     f.OmitEmpty,
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

func (f *_FieldInfo) GetValue(s *_StructInfo, structVal reflect.Value, opts *options) (reflect.Value, error) {
	parentVal, fieldVal, ok := s.FindGettableValue(structVal, f)
	if !ok {
		return reflect.Value{}, nil
	}
	var getter reflect.Value
	var methodInfo *_MethodInfo
	if len(f.PointerGetter) > 0 && parentVal.CanAddr() {
		parentVal = parentVal.Addr()
		if method, ok := parentVal.Type().MethodByName(f.PointerGetter); ok {
			methodInfo = &_MethodInfo{Method: method}
		}
	}
	if len(f.BaseGetter) > 0 {
		method, ok := parentVal.Type().MethodByName(f.BaseGetter)
		if ok {
			methodInfo = &_MethodInfo{Method: method}
		}
	}
	if methodInfo == nil {
		return fieldVal, nil
	}
	getter = parentVal.Method(methodInfo.Index)
	if !getter.IsValid() {
		return fieldVal, nil
	}
	return methodInfo.InvokeGetter(getter, opts)
}

func (f *_FieldInfo) SetValue(s *_StructInfo, structVal reflect.Value, opts *options, setFunc func(in reflect.Value) error) error {
	parentVal, fieldVal, ok := s.FindSettableValue(structVal, f)
	if !ok {
		return nil
	}
	//var setter reflect.Value
	var methodInfo *_MethodInfo
	if len(f.PointerSetter) > 0 && parentVal.CanAddr() {
		parentVal = parentVal.Addr()
		if method, ok := parentVal.Type().MethodByName(f.PointerSetter); ok {
			methodInfo = &_MethodInfo{Method: method}
		}
	}
	if len(f.BaseSetter) > 0 {
		if method, ok := parentVal.Type().MethodByName(f.BaseSetter); ok {
			methodInfo = &_MethodInfo{Method: method}
		}
	}
	if methodInfo == nil {
		return setFunc(fieldVal)
	}
	inVal := reflect.New(methodInfo.Type.In(methodInfo.Type.NumIn() - 1)).Elem()
	if err := setFunc(inVal); err != nil {
		return err
	}
	// 方法
	return methodInfo.InvokeSetter(parentVal, inVal, opts)
}

func (f *_FieldInfo) FuzzyMatch(field *_FieldInfo, opts *options, matchers ...func(field *_FieldInfo, opts *options) bool) int {
	var similarity int
	for i, matcher := range matchers {
		if matcher(field, opts) {
			similarity += len(matchers) - i
		}
	}
	return similarity
}

func (f *_FieldInfo) LabelsMatch(field *_FieldInfo, opts *options) bool {
	if len(f.Labels) != len(field.Labels) {
		return false
	}
	for i := range f.Labels {
		if !opts.EqualFold(f.Labels[i], field.Labels[i]) {
			return false
		}
	}
	return true
}

func (f *_FieldInfo) NamesMatch(field *_FieldInfo, opts *options) bool {
	if len(f.Names) != len(field.Names) {
		return false
	}
	for i := range f.Names {
		if !opts.EqualFold(f.Names[i], field.Names[i]) {
			return false
		}
	}
	return true
}

func (f *_FieldInfo) ParentMatch(field *_FieldInfo, opts *options) bool {
	return f.Parent.Type == field.Parent.Type
}

func (f *_FieldInfo) TypeMatch(field *_FieldInfo, opts *options) bool {
	return f.Type == field.Type
}

type _FieldInfos []*_FieldInfo

func (fs _FieldInfos) Sort(opts *options) {
	slices.SortFunc(fs, func(a, b *_FieldInfo) int {
		// 深度越浅越优先
		if len(a.Indexes) != len(b.Indexes) {
			return len(a.Indexes) - len(b.Indexes)
		}
		// 打上标签的优先
		if a.Tagged != b.Tagged {
			if a.Tagged {
				return -1
			} else {
				return 1
			}
		}
		// 标签名与字段名相同的优先
		if strings.EqualFold(a.Name, a.Label) {
			return -1
		}
		if strings.EqualFold(b.Name, b.Label) {
			return 1
		}
		// 字段index越小越优先
		for i, aIndex := range a.Indexes {
			bIndex := b.Indexes[i]
			if aIndex != bIndex {
				return aIndex - bIndex
			}
		}
		// 字段类型越基础有优先
		return _KindOrder[a.Type.Kind()] - _KindOrder[b.Type.Kind()]
	})
}

func (fs _FieldInfos) SortForFuzzyMatch(field *_FieldInfo, opts *options) {
	less := func(a, b *_FieldInfo) int {
		return b.FuzzyMatch(field, opts, b.LabelsMatch, b.NamesMatch, b.ParentMatch, b.TypeMatch) - a.FuzzyMatch(field, opts, a.LabelsMatch, a.NamesMatch, a.ParentMatch, a.TypeMatch)
	}
	slices.SortFunc(fs, less)
}

func (fs _FieldInfos) FindValueByField(field *_FieldInfo, opts *options) *_FieldInfo {
	tfs := slices.Clone(fs)
	tfs.SortForFuzzyMatch(field, opts)
	return tfs[0]
}

type _FieldIndexes map[string]_FieldInfos

func (fi _FieldIndexes) FindValueByLabel(label string, opts *options) *_FieldInfo {
	if fields, ok := fi[label]; ok {
		return fields[0]
	}
	for fieldLabel, fields := range fi {
		if !opts.EqualFold(fieldLabel, label) {
			continue
		}
		return fields[0]
	}
	return nil
}

func (fi _FieldIndexes) FindValueByField(field *_FieldInfo, opts *options) *_FieldInfo {
	if fields, ok := fi[field.Label]; ok {
		return fields.FindValueByField(field, opts)
	}
	for fieldLabel, fields := range fi {
		if !opts.EqualFold(fieldLabel, field.Label) {
			continue
		}
		return fields.FindValueByField(field, opts)
	}
	return nil
}

func (fi _FieldIndexes) Sort(opts *options) {
	for _, fs := range fi {
		fs.Sort(opts)
	}
}

type _MethodInfo struct {
	reflect.Method
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

func (m *_MethodInfo) InvokeSetter(receiverVal reflect.Value, inVal reflect.Value, opts *options) error {
	inValues := []reflect.Value{receiverVal}
	if m.Type.NumIn() == 3 {
		inValues = append(inValues, reflect.ValueOf(opts.Context))
	}
	inValues = append(inValues, inVal)
	outValues := m.Func.Call(inValues)
	if len(outValues) <= 0 {
		return nil
	}
	if err, ok := outValues[0].Interface().(error); ok && err != nil {
		return err
	}
	return nil
}

type _MethodInfos []*_MethodInfo

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
	val, _ := structCache.LoadOrStore(opts.TagKey, &sync.Map{})
	taggedStructCache := val.(*sync.Map)
	if f, ok := taggedStructCache.Load(typ); ok {
		return f.(*_StructInfo)
	}
	f, _ := taggedStructCache.LoadOrStore(typ, _NewStructInfo(typ).Analysis(opts))
	return f.(*_StructInfo)
}

var structCache sync.Map

type _MapEntry struct {
	KeyVal  reflect.Value
	Label   string
	ValVal  reflect.Value
	ValType reflect.Type
}

type _MapEntries []_MapEntry

func (s _MapEntries) Sort() {
	slices.SortFunc(s, func(a, b _MapEntry) int {
		if a.ValType.Kind() != b.ValType.Kind() {
			return _KindOrder[a.ValType.Kind()] - _KindOrder[b.ValType.Kind()]
		}
		return strings.Compare(a.Label, b.Label)
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
		if !valVal.IsValid() {
			continue
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
