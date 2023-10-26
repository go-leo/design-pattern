package main

import (
	"fmt"
	"github.com/go-leo/design-pattern/prototype/prototype"
)

func sampleArray() {
	//func main() {
	arr := []string{"1", "2", "3"}
	var arr2 []string
	err := prototype.Clone(&arr2, arr)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func number() {
	var num = 20
	var num2 int
	err := prototype.Clone(&num2, num)
	if err != nil {
		fmt.Println("error:", err)
	}
}

// func mainSample() {
func main() {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}

	type Toy struct {
		Model      string
		ColorGroup ColorGroup
		Materials  map[string]string
	}

	toy := Toy{
		Model: "bumblebee",
		ColorGroup: ColorGroup{
			ID:     1,
			Name:   "Reds",
			Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
		},
		Materials: map[string]string{
			"head": "iron",
			"body": "plastic",
		},
	}

	var clonedToy Toy
	err := prototype.Clone(&clonedToy, toy)
	if err != nil {
		fmt.Println("error:", err)
	}

}

func Unsupported() {
	type ColorGroup struct {
		ID      int
		Name    string
		Colors  []string
		Channel chan int
	}

	type Toy struct {
		Model      string
		ColorGroup ColorGroup
		Materials  map[string]string
	}

	toy := Toy{
		Model: "bumblebee",
		ColorGroup: ColorGroup{
			ID:     1,
			Name:   "Reds",
			Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
		},
		Materials: map[string]string{
			"head": "iron",
			"body": "plastic",
		},
	}

	err := prototype.Clone(nil, toy)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func SampleUnmarshal() {
	//func main() {
	var jsonBlob = []byte(`[
	{"Name": "Platypus", "Order": "Monotremata"},
	{"Name": "Quoll",    "Order": "Dasyuromorphia"}
]`)
	type Animal struct {
		Name  string
		Order string
	}
	var animals []Animal
	err := prototype.Unmarshal(jsonBlob, &animals)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v", animals)
}
