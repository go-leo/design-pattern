package prototype

type testFields struct {
	School string
}

type testGetter struct {
	testFields
	id      string
	name    string
	Age     int    `prototype:"age"`
	Address string `prototype:"address"`
}

func (t testGetter) Id() string {
	return "id:" + t.id
}

func (t testGetter) Name() (string, error) {
	return "name:" + t.name, nil
}

type testSetter struct {
	Id      string `prototype:"id"`
	Name    string `prototype:"name"`
	age     int
	address string
}

func (t *testSetter) SetAge(age int) {
	t.age = age * 2
}

func (t *testSetter) SetAddress(address string) {
	t.address = "china-" + address
}

type testGetterPointerParent struct {
	*testGetter
}

type testSetterPointerParent struct {
	*testSetter
}

//
//func TestStructInfo(t *testing.T) {
//	id := uuid.NewString()
//	name := "prototype"
//	age := 30
//	address := "shanghai"
//	src := testGetterPointerParent{
//		testGetter: &testGetter{
//			id:      id,
//			name:    name,
//			Age:     age,
//			Address: address,
//		},
//	}
//
//	opts := new(options).apply(TagKey("prototype")).correct()
//	info := _CachedStructInfo(reflect.TypeOf(src), opts)
//
//	t.Log(info)
//}
