//go:build solution

package errsqrt

import (
	"fmt"
	"math"
)

// ErrNegativeSqrt は負の数の平方根を求めようとしたことを表すエラーです。
type ErrNegativeSqrt float64

// Error は error インターフェースを実装します。
// fmt.Sprint(e) と書くと、fmt が e.Error() を呼び → また Sprint(e) … と無限再帰になるため、
// float64(e) に変換してから渡します。
func (e ErrNegativeSqrt) Error() string {
	return fmt.Sprint("cannot Sqrt negative number: ", float64(e))
}

// Sqrt は x の平方根を返します。x が負ならエラーを返します。
func Sqrt(x float64) (float64, error) {
	if x < 0 {
		return 0, ErrNegativeSqrt(x)
	}
	return math.Sqrt(x), nil
}
