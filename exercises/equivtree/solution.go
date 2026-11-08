//go:build solution

package equivtree

import "github.com/caffelatte23/golang-handson/exercises/equivtree/tree"

// Walk は t の値を昇順に ch へ送り、送り終えたら ch を close します。
func Walk(t *tree.Tree, ch chan int) {
	var walk func(*tree.Tree)
	walk = func(t *tree.Tree) {
		if t == nil {
			return
		}
		walk(t.Left)
		ch <- t.Value
		walk(t.Right)
	}
	walk(t)
	close(ch)
}

// Same は t1 と t2 が同じ値の列を持つかを返します。
func Same(t1, t2 *tree.Tree) bool {
	ch1, ch2 := make(chan int), make(chan int)
	go Walk(t1, ch1)
	go Walk(t2, ch2)
	for {
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2
		if ok1 != ok2 || v1 != v2 {
			// 途中で抜けると Walk の goroutine が送信でブロックしたまま残る (リーク)。
			// 残りを読み捨てて終了させる。
			go func() {
				for range ch1 {
				}
			}()
			go func() {
				for range ch2 {
				}
			}()
			return false
		}
		if !ok1 {
			return true
		}
	}
}
