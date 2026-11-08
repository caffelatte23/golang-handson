//go:build solution

package sqrt

import "math"

// Sqrt は x の平方根を返します (x >= 0 を前提とします)。
func Sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := 1.0
	for {
		next := z - (z*z-x)/(2*z)
		if math.Abs(next-z) < 1e-10 {
			return next
		}
		z = next
	}
}
