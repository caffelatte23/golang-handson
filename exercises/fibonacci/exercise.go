//go:build !solution

// Package fibonacci は第4章の演習です。
//
// 呼び出すたびにフィボナッチ数列 (0, 1, 1, 2, 3, 5, ...) の次の値を返す
// 関数 (クロージャ) を返す Fibonacci を実装してください。
//
// 確認: go test ./fibonacci/
package fibonacci

// Fibonacci は、呼ぶたびに次のフィボナッチ数を返す関数を返します。
func Fibonacci() func() int {
	// TODO: 外側の変数をクロージャで捕捉してください
	return func() int { return 0 }
}
