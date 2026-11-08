//go:build !solution

// Package generics は第6章の演習です。
//
//  1. Map: スライスの各要素に f を適用した新しいスライスを返す
//  2. Filter: keep が true を返す要素だけを残す
//  3. Reduce: 初期値 init から始めて f で畳み込む
//  4. List[T]: 単方向連結リスト。Push で末尾に追加、All で先頭から列挙 (iter.Seq[T])
//
// 確認: go test ./generics/
package generics

import "iter"

// Map は xs の各要素に f を適用したスライスを返します。
func Map[T, U any](xs []T, f func(T) U) []U {
	// TODO
	return nil
}

// Filter は keep が true を返す要素だけを残したスライスを返します。
func Filter[T any](xs []T, keep func(T) bool) []T {
	// TODO
	return nil
}

// Reduce は xs を f で畳み込みます。
func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	// TODO
	return init
}

// List は単方向連結リストです。ゼロ値で使えるようにしてください。
type List[T any] struct {
	// TODO: head, tail などのフィールド
}

// Push は v を末尾に追加します。
func (l *List[T]) Push(v T) {
	// TODO
}

// All は先頭から順に要素を返すイテレータです。
func (l *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		// TODO
	}
}
