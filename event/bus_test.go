package event

import (
	"sync"
	"testing"
)

func TestMap(t *testing.T) {
	var num = []int{1, 2, 3, 4, 5}
	var seq = []int{0, 9, 8, 7, 6}

	var m sync.Map
	m.Store("slice", &num)
	t.Log(m.Load("slice"))
	t.Log(m.CompareAndSwap("slice", &num, &seq))
	t.Log(m.Load("slice"))
}

func TestMap_CompareAndSwap(t *testing.T) {
	var num = []int{1, 2, 3, 4, 5}
	var seq = []int{0, 9, 8, 7, 6}

	var m sync.Map
	m.CompareAndSwap("slice", nil, &num)
	t.Log(m.Load("slice"))
	t.Log(m.CompareAndSwap("slice", &num, &seq))
	t.Log(m.Load("slice"))
}
