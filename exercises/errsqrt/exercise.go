//go:build !solution

// Package errsqrt は第7章の演習です。
//
// 負の数を渡されたらエラーを返す Sqrt を実装してください。
//
//  1. ErrNegativeSqrt 型 (float64 ベース) に Error() string を実装し、
//     "cannot Sqrt negative number: -2" のようなメッセージを返す
//  2. Sqrt は x < 0 のとき ErrNegativeSqrt(x) をエラーとして返す
//
// ヒント: Error メソッドの中で fmt.Sprint(e) を呼ぶと無限再帰になり、スタックオーバーフローで異常終了します。なぜでしょう?
//
// 確認: go test ./errsqrt/
package errsqrt

// ErrNegativeSqrt は負の数の平方根を求めようとしたことを表すエラーです。
type ErrNegativeSqrt float64

// TODO: func (e ErrNegativeSqrt) Error() string

// Sqrt は x の平方根を返します。x が負ならエラーを返します。
func Sqrt(x float64) (float64, error) {
	// TODO
	return 0, nil
}
