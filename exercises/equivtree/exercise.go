//go:build !solution

// Package equivtree は第8章の演習です。
//
// 同じ値の集合を持つ二分探索木でも、形はさまざまです。
// goroutine と channel を使って、2 つの木が同じ値の列を持つか判定してください。
//
//  1. Walk(t, ch): 木 t の値を小さい順に ch へ送り、最後に ch を close する
//  2. Same(t1, t2): 2 つの木を並行に Walk し、値の列が等しいかを返す
//
// 確認: go test -race ./equivtree/
package equivtree

import "github.com/caffelatte23/golang-handson/exercises/equivtree/tree"

// Walk は t の値を昇順に ch へ送り、送り終えたら ch を close します。
func Walk(t *tree.Tree, ch chan int) {
	// TODO: 再帰用の内部関数を作り、最後に close(ch) する
	close(ch)
}

// Same は t1 と t2 が同じ値の列を持つかを返します。
func Same(t1, t2 *tree.Tree) bool {
	// TODO
	return false
}
