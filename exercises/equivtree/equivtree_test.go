package equivtree

import (
	"slices"
	"testing"

	"github.com/caffelatte23/golang-hanson/exercises/equivtree/tree"
)

func TestWalk(t *testing.T) {
	ch := make(chan int)
	go Walk(tree.New(1), ch)
	var got []int
	for v := range ch { // Walk が close しないとここで永久に待つ
		got = append(got, v)
	}
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if !slices.Equal(got, want) {
		t.Errorf("Walk = %v, want %v", got, want)
	}
}

func TestSame(t *testing.T) {
	if !Same(tree.New(1), tree.New(1)) {
		t.Error("Same(New(1), New(1)) = false, want true")
	}
	if Same(tree.New(1), tree.New(2)) {
		t.Error("Same(New(1), New(2)) = true, want false")
	}
}
