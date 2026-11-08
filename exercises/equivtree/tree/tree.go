// Package tree は演習用の二分探索木を提供します (演習の対象ではありません)。
package tree

import (
	"fmt"
	"math/rand/v2"
)

// Tree は整数を保持する二分木のノードです。
type Tree struct {
	Left  *Tree
	Value int
	Right *Tree
}

// New は k, 2k, ..., 10k を値に持つ、ランダムな形の二分探索木を返します。
func New(k int) *Tree {
	var t *Tree
	for _, v := range rand.Perm(10) {
		t = insert(t, (1+v)*k)
	}
	return t
}

func insert(t *Tree, v int) *Tree {
	if t == nil {
		return &Tree{nil, v, nil}
	}
	if v < t.Value {
		t.Left = insert(t.Left, v)
	} else {
		t.Right = insert(t.Right, v)
	}
	return t
}

func (t *Tree) String() string {
	if t == nil {
		return "()"
	}
	s := ""
	if t.Left != nil {
		s += t.Left.String() + " "
	}
	s += fmt.Sprint(t.Value)
	if t.Right != nil {
		s += " " + t.Right.String()
	}
	return "(" + s + ")"
}
