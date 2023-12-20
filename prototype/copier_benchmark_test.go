package prototype_test

import (
	"github.com/go-leo/design-pattern/prototype"
	"testing"
	"time"
)

type User struct {
	Name     string
	Birthday *time.Time
	Nickname string
	Role     string
	Age      int32
	FakeAge  *int32
	Notes    []string
	flags    []byte
}

type Employee struct {
	_User     *User
	Name      string
	Birthday  *time.Time
	NickName  *string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []*string
	flags     []byte
}

func BenchmarkCopyStruct(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
	for x := 0; x < b.N; x++ {
		prototype.Clone(&Employee{}, &user)
	}
}

func BenchmarkCopyStructFields(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
	for x := 0; x < b.N; x++ {
		prototype.Clone(&Employee{}, &user)
	}
}

//
//func BenchmarkNamaCopy(b *testing.B) {
//	var fakeAge int32 = 12
//	user := User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
//	for x := 0; x < b.N; x++ {
//		employee := &Employee{
//			Name:      user.Name,
//			NickName:  &user.Nickname,
//			Age:       int64(user.Age),
//			FakeAge:   int(*user.FakeAge),
//			DoubleAge: user.DoubleAge(),
//		}
//
//		for _, note := range user.Notes {
//			employee.Notes = append(employee.Notes, &note)
//		}
//		employee.Role(user.Role)
//	}
//}

//func BenchmarkJsonMarshalCopy(b *testing.B) {
//	var fakeAge int32 = 12
//	user := User{Name: "Jinzhu", Nickname: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, flags: []byte{'x'}}
//	for x := 0; x < b.N; x++ {
//		data, _ := json.Marshal(user)
//		var employee Employee
//		json.Unmarshal(data, &employee)
//
//		employee.DoubleAge = user.DoubleAge()
//		employee.Role(user.Role)
//	}
//}
