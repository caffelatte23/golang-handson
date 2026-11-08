package generics

import (
	"slices"
	"strconv"
	"testing"
)

func TestMapFilterReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	strs := Map(nums, strconv.Itoa)
	if want := []string{"1", "2", "3", "4", "5"}; !slices.Equal(strs, want) {
		t.Errorf("Map = %v, want %v", strs, want)
	}
	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	if want := []int{2, 4}; !slices.Equal(evens, want) {
		t.Errorf("Filter = %v, want %v", evens, want)
	}
	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	if sum != 15 {
		t.Errorf("Reduce(sum) = %d, want 15", sum)
	}
	joined := Reduce(strs, "", func(acc, s string) string { return acc + s })
	if joined != "12345" {
		t.Errorf("Reduce(join) = %q", joined)
	}
}

func TestList(t *testing.T) {
	var l List[string] // ゼロ値で使える
	if got := slices.Collect(l.All()); len(got) != 0 {
		t.Errorf("empty list = %v", got)
	}
	for _, s := range []string{"a", "b", "c"} {
		l.Push(s)
	}
	if got, want := slices.Collect(l.All()), []string{"a", "b", "c"}; !slices.Equal(got, want) {
		t.Errorf("All = %v, want %v", got, want)
	}
	// break で途中終了しても panic しない
	for v := range l.All() {
		if v == "b" {
			break
		}
	}
}
