package prototype

import (
	"github.com/go-leo/design-pattern/prototype/internal"
	"golang.org/x/exp/slices"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// A field represents a single field found in a struct.
type jfield struct {
	name       string
	tag        bool
	index      []int
	typ        reflect.Type
	clonerFunc clonerFunc
	fullName   string
}

type jstructFields struct {
	// dominants 是一个字段列表，存储了结构体的主要字段信息
	dominants []jfield
	// dominantsNameIndex 是一个映射，用于通过字段名称查找字段在 dominants 中的索引
	dominantsNameIndex      map[string]int
	recessives              []jfield
	recessivesNameIndex     map[string][]int
	recessivesFullNameIndex map[string]int
	selfFields              []jfield
	selfNameIndex           map[string]int
}

// byIndex sorts field by index sequence.
type byIndex []jfield

func (x byIndex) Len() int { return len(x) }

func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

var fieldCache sync.Map // map[reflect.Type]structFields

// typeFields 函数返回给定类型应该被识别的字段列表。
// 该算法是对要包含的结构体集合进行广度优先搜索 - 首先是顶级结构体，然后是任何可达的匿名结构体。
// 简单来说，typeFields 函数用于获取应该处理的字段列表。
// 它使用广度优先搜索算法遍历结构体类型，包括顶级结构体和可达的匿名结构体，并返回这些结构体中应该被处理的字段列表。
func typeFields(t reflect.Type, opts *options) jstructFields {
	// current 和 next 两个用于存储当前和下一级的匿名字段的切片
	current := make([]jfield, 0)
	next := []jfield{{typ: t}}

	// currentCount 和 nextCount 用于记录字段名称出现的次数
	var currentCount, nextCount map[reflect.Type]int

	// visited 用于记录已经访问过的类型
	visited := map[reflect.Type]bool{}

	// fields 用于存储找到的字段
	var fields []jfield

	var selfFields []jfield

	// len(next) > 0，表示还有下一级的匿名字段需要探索
	for len(next) > 0 {
		// 1. 首先交换 current 和 next，并清空 next 切片。 交换是为了将当前级别的匿名字段作为下一级的匿名字段进行探索
		current, next = next, current[:0]
		currentCount, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			// 2. 对于每个字段 f，首先检查是否已经访问过该字段的类型。
			// 如果已经访问过，则跳过。
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// 3. 遍历字段 f 的类型的每个字段 sf。
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				// 4. 对于每个字段 sf，根据一定的规则判断是否需要包含该字段, 规则如下
				//   - 如果 sf 是匿名字段，则检查它的类型是否是导出的结构体类型，如果不是则忽略。
				//   - 如果 sf 不是匿名字段且不是导出的字段，则忽略。
				if sf.Anonymous {
					// 对于匿名字段，检查字段类型是否为指针类型，如果是，则将类型设置为指针所指向的类型。
					// 如果字段是未导出的非结构体类型，则忽略该字段。
					t := sf.Type
					if t.Kind() == reflect.Pointer {
						t = t.Elem()
					}
					if !sf.IsExported() && t.Kind() != reflect.Struct {
						continue
					}
					// 在处理结构体字段时，不要忽略未导出的结构体类型的嵌入字段，因为这些嵌入字段可能具有导出的字段。
				} else if !sf.IsExported() {
					// 对于非匿名字段，如果字段未导出，则忽略该字段。
					continue
				}

				// 5. 接下来，它会获取字段的标签，并解析标签中的名称和选项。
				name, ok := tagName(sf, opts.TagKey)
				if !ok {
					continue
				}

				// 8. 复制字段的索引序列，并将当前字段的索引添加到该序列中。
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type

				// 9. 如果字段的类型是指针类型且没有名称，则将类型设置为指针所指向的类型。
				if ft.Name() == "" && ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
				}

				tagged := name != ""
				if name == "" {
					name = sf.Name
				}
				currField := jfield{
					name:       name,
					tag:        tagged,
					index:      index,
					typ:        ft,
					clonerFunc: typeCloner(typeByIndex(t, index), true, opts),
					fullName:   fullName(f, name, sf),
				}
				if t == f.typ {
					selfFields = append(selfFields, currField)
				}
				// 10. 记录找到的字段信息，并根据字段所属类型的计数决定是否添加多个副本。
				if tagged || !sf.Anonymous || ft.Kind() != reflect.Struct {
					// 记录字段路径
					fields = append(fields, currField)
					if currentCount[f.typ] > 1 {
						// 如果有多个实例，添加第二个，这样湮灭代码将看到一个副本。
						// 它只关心1和2之间的区别，所以不要再生成任何副本了。
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// 11. 如果字段是匿名结构体，则记录该结构体以便在下一轮中继续探索。
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, currField)
				}
			}
		}
	}

	// 对字段进行排序，首先按照名称排序，然后按照字段索引长度排序，最后按照是否有标签排序。
	sort.Slice(fields, func(i, j int) bool {
		x := fields
		if len(x[i].index) != len(x[j].index) {
			return len(x[i].index) < len(x[j].index)
		}
		if x[i].tag != x[j].tag {
			return x[i].tag
		}
		return byIndex(x).Less(i, j)
	})

	selfFields, selfNameIndex := selfNameIndex(t, opts, selfFields)

	// 区分出主要字段和次要字段
	dominants, recessives := divideFields(fields)
	dominants, dominantsNameIndex := dominantsNameIndex(t, opts, dominants)
	recessives, recessivesNameIndex, recessivesFullnameIndex := recessivesNameIndex(t, opts, recessives)
	return jstructFields{
		dominants:               dominants,
		dominantsNameIndex:      dominantsNameIndex,
		recessives:              recessives,
		recessivesNameIndex:     recessivesNameIndex,
		recessivesFullNameIndex: recessivesFullnameIndex,
		selfFields:              selfFields,
		selfNameIndex:           selfNameIndex,
	}
}

func selfNameIndex(t reflect.Type, opts *options, fields []jfield) ([]jfield, map[string]int) {
	// 对字段进行排序，按照索引顺序排序。
	sort.Sort(byIndex(fields))
	// 创建一个映射 nameIndex，用于通过字段名称查找字段在 fields 中的索引。
	nameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = i
	}
	return fields, nameIndex
}

// divideFields 分出主要字段和次要字段
func divideFields(fields []jfield) ([]jfield, []jfield) {
	dominants := make([]jfield, 0, len(fields))
	recessives := make([]jfield, 0)
	for advance, i := 0, 0; i < len(fields); i += advance {
		// 进行循环，每次循环处理一个字段名称。 在循环内部，查找具有相同名称的字段序列。
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 {
			// 如果只有一个字段具有该名称，则将该字段添加到输出切片中。
			dominants = append(dominants, fi)
			continue
		}
		group := fields[i : i+advance]
		if len(group) > 1 && len(group[0].index) == len(group[1].index) && group[0].tag == group[1].tag {
			// 如果有多个name相同，又在同一级，则一个结构体有两个相同的字段，这种情况全部都忽略
			continue
		}
		dominants = append(dominants, group[0])
		recessives = append(recessives, group[1:]...)
	}
	return dominants, recessives
}

func recessivesNameIndex(t reflect.Type, opts *options, fields []jfield) ([]jfield, map[string][]int, map[string]int) {
	// 对字段进行排序，按照索引顺序排序。
	sort.Sort(byIndex(fields))
	// 创建一个映射 nameIndex，用于通过字段名称查找字段在 fields 中的索引。
	nameIndex := make(map[string][]int, len(fields))
	fullNameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = append(nameIndex[field.name], i)
		fullNameIndex[field.fullName] = i
	}
	return fields, nameIndex, fullNameIndex
}

func dominantsNameIndex(t reflect.Type, opts *options, fields []jfield) ([]jfield, map[string]int) {
	// 对字段进行排序，按照索引顺序排序。
	sort.Sort(byIndex(fields))
	// 创建一个映射 nameIndex，用于通过字段名称查找字段在 fields 中的索引。
	nameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = i
	}
	return fields, nameIndex
}

func tagName(sf reflect.StructField, Key string) (string, bool) {
	tagVal := sf.Tag.Get(Key)
	// 如果是标签是"-",则忽略该字段
	if tagVal == "-" {
		return "", false
	}
	tag, opt, _ := strings.Cut(tagVal, ",")
	// 获取字段的标签，并解析标签中的名称和选项。
	name, opts := tag, tagOptions(opt)

	// 忽略其他 tag options
	_ = opts
	return name, true
}

func fullName(f jfield, name string, sf reflect.StructField) string {
	var fullName string
	if len(f.fullName) > 0 {
		fullName = strings.Join([]string{f.fullName, name}, ".")
	} else {
		fullName = sf.Name
	}
	return fullName
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

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

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

		// 有tag的字段
		if field.WithTag {
			s.FieldIndexes[field.Label] = append(s.FieldIndexes[field.Label], field)
		}

		// 导出字段:
		// 所有导出(包括非匿名和匿名)字段
		if field.IsExported() {
			s.FieldIndexes[field.Label] = append(s.FieldIndexes[field.Label], field)
			continue
		}

		// 非导出字段：
		// 非导出、匿名结构体字段或者匿名结构体指针字段,
		if field.Type.Kind() == reflect.Struct ||
			field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct {
			s.FieldIndexes[field.Label] = append(s.FieldIndexes[field.Label], field)
			continue
		}
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
		s.StructMethodIndexes[method.Name] = &methodInfo{Method: method}
	}
	ptrType := reflect.PointerTo(typ)
	for i := 0; i < ptrType.NumMethod(); i++ {
		method := ptrType.Method(i)
		// 忽略未导出方法
		if !method.IsExported() {
			continue
		}
		s.PointerMethodIndexes[method.Name] = &methodInfo{Method: method}
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

func (s *structInfo) AnalysisField(field *fieldInfo, opts *options) {

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

func cachedStruct(t reflect.Type, opts *options) *structInfo {
	if f, ok := fieldCache.Load(t); ok {
		return f.(*structInfo)
	}
	f, _ := fieldCache.LoadOrStore(t, newStructInfo(t).Analysis(opts))
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
