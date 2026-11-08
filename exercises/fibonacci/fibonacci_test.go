package fibonacci

import (
	"slices"
	"testing"
)

func TestFibonacci(t *testing.T) {
	f := Fibonacci()
	var got []int
	for range 10 {
		got = append(got, f())
	}
	want := []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	// 別のクロージャは独立した状態を持つ
	g := Fibonacci()
	if v := g(); v != 0 {
		t.Errorf("new closure first value = %d, want 0", v)
	}
}
