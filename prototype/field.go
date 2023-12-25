package prototype

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// A field represents a single field found in a struct.
type field struct {
	name       string
	tag        bool
	index      []int
	typ        reflect.Type
	clonerFunc clonerFunc
	fullName   string
}

type jstructFields struct {
	// dominants 是一个字段列表，存储了结构体的主要字段信息
	dominants []field
	// dominantsNameIndex 是一个映射，用于通过字段名称查找字段在 dominants 中的索引
	dominantsNameIndex      map[string]int
	recessives              []field
	recessivesNameIndex     map[string][]int
	recessivesFullNameIndex map[string]int
	selfFields              []field
	selfNameIndex           map[string]int
}

// byIndex sorts field by index sequence.
type byIndex []field

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

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type, opts *options, tagKey string) jstructFields {
	key := t.String() + ":" + tagKey
	if f, ok := fieldCache.Load(key); ok {
		return f.(jstructFields)
	}
	fields := typeFields(t, opts)

	f, _ := fieldCache.LoadOrStore(key, fields)
	return f.(jstructFields)
}

// typeFields 函数返回给定类型应该被识别的字段列表。
// 该算法是对要包含的结构体集合进行广度优先搜索 - 首先是顶级结构体，然后是任何可达的匿名结构体。
// 简单来说，typeFields 函数用于获取应该处理的字段列表。
// 它使用广度优先搜索算法遍历结构体类型，包括顶级结构体和可达的匿名结构体，并返回这些结构体中应该被处理的字段列表。
func typeFields(t reflect.Type, opts *options) jstructFields {
	// current 和 next 两个用于存储当前和下一级的匿名字段的切片
	current := make([]field, 0)
	next := []field{{typ: t}}

	// currentCount 和 nextCount 用于记录字段名称出现的次数
	var currentCount, nextCount map[reflect.Type]int

	// visited 用于记录已经访问过的类型
	visited := map[reflect.Type]bool{}

	// fields 用于存储找到的字段
	var fields []field

	var selfFields []field

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
				currField := field{
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
		if x[i].name != x[j].name {
			return x[i].name < x[j].name
		}
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

func selfNameIndex(t reflect.Type, opts *options, fields []field) ([]field, map[string]int) {
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
func divideFields(fields []field) ([]field, []field) {
	dominants := make([]field, 0, len(fields))
	recessives := make([]field, 0)
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

func recessivesNameIndex(t reflect.Type, opts *options, fields []field) ([]field, map[string][]int, map[string]int) {
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

func dominantsNameIndex(t reflect.Type, opts *options, fields []field) ([]field, map[string]int) {
	// 对字段进行排序，按照索引顺序排序。
	sort.Sort(byIndex(fields))
	// 创建一个映射 nameIndex，用于通过字段名称查找字段在 fields 中的索引。
	nameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = i
	}
	return fields, nameIndex
}

func findDominantField(tgtFields jstructFields, opts *options, tagName string) (field, bool) {
	return findField(tgtFields.dominantsNameIndex, tgtFields.dominants, opts, tagName)
}

func findField(nameIndex map[string]int, fields []field, opts *options, tagName string) (field, bool) {
	// 查找tgt字段
	if tgtIdx, ok := nameIndex[tagName]; ok {
		// 找到了一个完全匹配的字段名称
		return fields[tgtIdx], true
	} else {
		// 代码回退到了一种更为耗时的线性搜索方法，该方法在进行字段名称匹配时不考虑大小写
		for tgtKey, tgtIdx := range nameIndex {
			if opts.NameComparer(tgtKey, tagName) {
				return fields[tgtIdx], true
			}
		}
	}
	return field{}, false
}

func findRecessiveField(tgtFields jstructFields, opts *options, tagKey string) ([]field, bool) {
	// 查找tgt字段
	if tgtIdxs, ok := tgtFields.recessivesNameIndex[tagKey]; ok {
		// 找到了一个完全匹配的字段名称
		fields := make([]field, 0, len(tgtIdxs))
		for _, tgtIdx := range tgtIdxs {
			fields = append(fields, tgtFields.recessives[tgtIdx])
		}
		return fields, true
	} else {
		// 代码回退到了一种更为耗时的线性搜索方法，该方法在进行字段名称匹配时不考虑大小写
		for tgtKey, tgtIdxs := range tgtFields.recessivesNameIndex {
			if opts.NameComparer(tgtKey, tagKey) {
				fields := make([]field, 0, len(tgtIdxs))
				for _, tgtIdx := range tgtIdxs {
					fields = append(fields, tgtFields.recessives[tgtIdx])
				}
				if len(fields) > 0 {
					return fields, true
				}
			}

		}
	}
	return nil, false
}

func findValue(val reflect.Value, f field) (reflect.Value, bool) {
	outVal := val
	for _, i := range f.index {
		if outVal.Kind() == reflect.Pointer {
			if outVal.IsNil() {
				return reflect.Value{}, false
			}
			outVal = outVal.Elem()
		}
		outVal = outVal.Field(i)
	}
	return outVal, true
}

func findSettableValue(val reflect.Value, f field) (reflect.Value, error) {
	outVal := val
	for _, i := range f.index {
		if outVal.Kind() == reflect.Pointer {
			if outVal.IsNil() {
				if !outVal.CanSet() {
					return reflect.Value{}, fmt.Errorf("prototype: cannot set embedded pointer to unexported struct: %v", outVal.Type().Elem())
				}
				outVal.Set(reflect.New(outVal.Type().Elem()))
			}
			outVal = outVal.Elem()
		}
		outVal = outVal.Field(i)
	}
	return outVal, nil
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
	if !isValidTag(name) {
		// 如果名称无效，则将名称设置为空。
		name = ""
	}
	// 忽略其他 tag options
	_ = opts
	return name, true
}

func fullName(f field, name string, sf reflect.StructField) string {
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

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag Nil.
		case !unicode.IsLetter(c) && !unicode.IsDigit(c):
			return false
		}
	}
	return true
}

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// Contains reports whether a comma-separated dominants of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var name string
		name, s, _ = strings.Cut(s, ",")
		if name == optionName {
			return true
		}
	}
	return false
}
