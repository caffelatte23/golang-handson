//go:build solution

package pic

// Pic は dy 行 dx 列の 2 次元スライスを返します。値は x^y (XOR) にしてください。
func Pic(dx, dy int) [][]uint8 {
	p := make([][]uint8, dy)
	for y := range p {
		p[y] = make([]uint8, dx)
		for x := range p[y] {
			p[y][x] = uint8(x ^ y)
		}
	}
	return p
}
