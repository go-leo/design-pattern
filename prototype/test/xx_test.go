package test

import (
	"golang.org/x/exp/slices"
	"testing"
)

func TestSort(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6}
	slices.SortFunc(a, func(a, b int) bool {
		return a < b
	})
	t.Log(a)

	a = []int{1, 2, 3, 4, 5, 6}
	slices.SortFunc(a, func(a, b int) bool {
		return a > b
	})
	t.Log(a)
}
