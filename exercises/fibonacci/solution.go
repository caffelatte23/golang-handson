//go:build solution

package fibonacci

// Fibonacci は、呼ぶたびに次のフィボナッチ数を返す関数を返します。
func Fibonacci() func() int {
	a, b := 0, 1
	return func() int {
		v := a
		a, b = b, a+b // 多重代入で一時変数が要らない
		return v
	}
}
