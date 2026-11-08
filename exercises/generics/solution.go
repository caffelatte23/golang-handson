//go:build solution

package generics

import "iter"

// Map は xs の各要素に f を適用したスライスを返します。
func Map[T, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))
	for _, x := range xs {
		out = append(out, f(x))
	}
	return out
}

// Filter は keep が true を返す要素だけを残したスライスを返します。
func Filter[T any](xs []T, keep func(T) bool) []T {
	var out []T
	for _, x := range xs {
		if keep(x) {
			out = append(out, x)
		}
	}
	return out
}

// Reduce は xs を f で畳み込みます。
func Reduce[T, A any](xs []T, init A, f func(A, T) A) A {
	acc := init
	for _, x := range xs {
		acc = f(acc, x)
	}
	return acc
}

type node[T any] struct {
	val  T
	next *node[T]
}

// List は単方向連結リストです。ゼロ値で使えるようにしてください。
type List[T any] struct {
	head, tail *node[T]
}

// Push は v を末尾に追加します。
func (l *List[T]) Push(v T) {
	n := &node[T]{val: v}
	if l.tail == nil {
		l.head, l.tail = n, n
		return
	}
	l.tail.next = n
	l.tail = n
}

// All は先頭から順に要素を返すイテレータです。
func (l *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for n := l.head; n != nil; n = n.next {
			if !yield(n.val) {
				return
			}
		}
	}
}
