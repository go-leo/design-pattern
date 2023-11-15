package prototype

import (
	"github.com/go-leo/gox/stringx"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// A field represents a single field found in a struct.
type field struct {
	name       string
	nameBytes  []byte                 // []byte(name)
	equalFold  func(s, t []byte) bool // bytes.EqualFold or equivalent
	tag        bool
	index      []int
	typ        reflect.Type
	clonerFunc ClonerFunc
}

type structFields struct {
	// list 是一个字段列表，存储了结构体的字段信息
	list []field
	// nameIndex 是一个映射，用于通过字段名称查找字段在 list 中的索引
	nameIndex map[string]int
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

// dominantField looks through the fields, all of which are known to
// have the same Nil, to find the single field that dominates the
// others using Go's embedding rules, modified by the presence of
// JSON tags. If there are multiple top-level fields, the boolean
// will be false: This condition is an error in Go and we skip all
// the fields.
func dominantField(fields []field) (field, bool) {
	// The fields are sorted in increasing index-length order, then by presence of tag.
	// That means that the first field is the dominant one. We need only check
	// for error cases: two fields at top level, either both tagged or neither tagged.
	if len(fields) > 1 && len(fields[0].index) == len(fields[1].index) && fields[0].tag == fields[1].tag {
		return field{}, false
	}
	return fields[0], true
}

var fieldCache sync.Map // map[reflect.Type]structFields

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type, opts *options, isSrc bool) structFields {
	key := t.String() + ":" + strconv.FormatBool(isSrc)
	if f, ok := fieldCache.Load(key); ok {
		return f.(structFields)
	}
	fields := typeFields(t, opts, isSrc)
	f, _ := fieldCache.LoadOrStore(key, fields)
	return f.(structFields)
}

// typeFields 函数返回给定类型应该被识别的字段列表。
// 该算法是对要包含的结构体集合进行广度优先搜索 - 首先是顶级结构体，然后是任何可达的匿名结构体。
// 简单来说，typeFields 函数用于获取应该处理的字段列表。
// 它使用广度优先搜索算法遍历结构体类型，包括顶级结构体和可达的匿名结构体，并返回这些结构体中应该被处理的字段列表。
func typeFields(t reflect.Type, opts *options, isSrc bool) structFields {
	// current 和 next 两个用于存储当前和下一级的匿名字段的切片
	current := []field{}
	next := []field{{typ: t}}

	// currentCount 和 nextCount 用于记录字段名称出现的次数
	var currentCount, nextCount map[reflect.Type]int

	// visited 用于记录已经访问过的类型
	visited := map[reflect.Type]bool{}

	// fields 用于存储找到的字段
	var fields []field

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
				var tagKey string
				if isSrc {
					tagKey = opts.SourceTagKey
				} else {
					tagKey = opts.TargetTagKey
				}
				tag := sf.Tag.Get(tagKey)

				// 6. 如果是标签是"-",则忽略该字段
				if tag == "-" {
					continue
				}

				// 7. 获取字段的标签，并解析标签中的名称和选项。
				name, opts := parseTag(tag)
				if !isValidTag(name) {
					// 如果名称无效，则将名称设置为空。
					name = ""
				}
				// 忽略其他 tag options
				_ = opts

				// 8. 复制字段的索引序列，并将当前字段的索引添加到该序列中。
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type

				// 9. 如果字段的类型是指针类型且没有名称，则将类型设置为指针所指向的类型。
				if ft.Name() == "" && ft.Kind() == reflect.Pointer {
					ft = ft.Elem()
				}

				// Record found field and index sequence.
				// 10. 记录找到的字段信息，并根据字段所属类型的计数决定是否添加多个副本。
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					field := field{
						name:  name,
						tag:   tagged,
						index: index,
						typ:   ft,
						//omitEmpty: opts.Contains("omitempty"),
					}
					field.nameBytes = []byte(field.name)
					field.equalFold = stringx.FoldFunc(field.nameBytes)

					fields = append(fields, field)
					if currentCount[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// 11. 如果字段是匿名结构体，则记录该结构体以便在下一轮中继续探索。
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, field{name: ft.Name(), index: index, typ: ft})
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

	// 删除被 Go 语言规则隐藏的字段，只保留具有相同名称的字段中的一个。

	// 对字段进行排序，并根据一定的规则选择主要字段来处理隐藏字段。
	// 这样做可以确保在后续的处理中，只保留了具有相同名称的字段中的一个主要字段。

	// 具体来说，对于具有相同名称的字段序列，代码会选择一个主要字段来保留，而删除其他隐藏字段。
	// 选择主要字段的规则是调用 dominantField 函数，该函数会根据一定的规则选择一个字段作为主要字段。
	// 主要字段是指在隐藏字段中具有较高优先级的字段。

	// 创建一个空的切片 out，用于存储经过隐藏字段处理后的字段信息。
	//out := fields[:0]
	//for advance, i := 0, 0; i < len(fields); i += advance {
	//	// 进行循环，每次循环处理一个字段名称。 在循环内部，查找具有相同名称的字段序列。
	//
	//	fi := fields[i]
	//	name := fi.name
	//	for advance = 1; i+advance < len(fields); advance++ {
	//		fj := fields[i+advance]
	//		if fj.name != name {
	//			break
	//		}
	//	}
	//	if advance == 1 {
	//		// 如果只有一个字段具有该名称，则将该字段添加到输出切片中。
	//		out = append(out, fi)
	//		continue
	//	}
	//	dominant, ok := dominantField(fields[i : i+advance])
	//	if ok {
	//		// 如果有多个字段具有相同的名称，则使用 dominantField 函数找到其中的一个主要字段，并将其添加到输出切片中。
	//		out = append(out, dominant)
	//	}
	//}
	//
	//// 将处理后的字段信息赋值给 fields
	//fields = out

	// 对字段进行排序，按照索引顺序排序。
	sort.Sort(byIndex(fields))

	// 对每个字段，设置其克隆器cloner
	for i := range fields {
		f := &fields[i]
		f.clonerFunc = typeCloner(typeByIndex(t, f.index), opts)
	}

	// 创建一个映射 nameIndex，用于通过字段名称查找字段在 fields 中的索引。
	nameIndex := make(map[string]int, len(fields))
	for i, field := range fields {
		nameIndex[field.name] = i
	}

	// 返回包含字段信息和字段索引映射的 structFields 结构体。
	return structFields{fields, nameIndex}
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
