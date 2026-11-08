package equivtree

import (
	"runtime"
	"slices"
	"testing"
	"time"

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

// build は値を順に挿入した二分探索木を作ります (形を制御するためのテスト用ヘルパー)。
func build(vals ...int) *tree.Tree {
	var root *tree.Tree
	for _, v := range vals {
		root = insert(root, v)
	}
	return root
}

func insert(t *tree.Tree, v int) *tree.Tree {
	if t == nil {
		return &tree.Tree{Value: v}
	}
	if v < t.Value {
		t.Left = insert(t.Left, v)
	} else {
		t.Right = insert(t.Right, v)
	}
	return t
}

func TestSame(t *testing.T) {
	before := runtime.NumGoroutine()

	tests := []struct {
		name   string
		t1, t2 *tree.Tree
		want   bool
	}{
		{"同じ値・ランダムな形", tree.New(1), tree.New(1), true},
		{"値がすべて違う", tree.New(1), tree.New(2), false},
		{"同じ値・違う形", build(1, 2, 3, 4, 5), build(3, 1, 4, 2, 5), true},
		{"中央の値だけ違う", build(1, 2, 3, 4, 5), build(1, 2, 9, 4, 5), false},
		{"合計は同じだが値が違う", build(1, 4), build(2, 3), false},
		{"片方が長い", build(1, 2, 3), build(1, 2, 3, 4), false},
		{"片方が短い", build(1, 2, 3, 4), build(1, 2, 3), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Same(tt.t1, tt.t2); got != tt.want {
				t.Errorf("Same = %v, want %v", got, tt.want)
			}
		})
	}

	// 途中で false を返しても、Walk の goroutine が残り続けない (リークしない) こと
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Errorf("goroutine leak: %d goroutines before, %d after", before, n)
	}
}
