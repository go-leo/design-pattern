package main

import "math"

func main() {
	Num(0)
	NumFunc(0, func(x int) int {
		return x
	})
}

func Num(x int) int {
	if x == math.MaxInt64 {
		return x
	}
	x++
	return Num(x)
}

func NumFunc(x int, f func(x int) int) func(x int) int {
	if x == math.MaxInt64 {
		return func(_ int) int {
			return x
		}
	}
	x++
	f2 := func(x int) int {
		return NumFunc(x, func(x int) int {
			return x
		})(x)
	}
	return f2
}
