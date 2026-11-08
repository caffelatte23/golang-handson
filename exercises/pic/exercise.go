//go:build !solution

// Package pic は第4章の演習です。
//
// 長さ dy のスライスで、各要素が長さ dx の []uint8 であるものを返す Pic を実装してください。
// 各ピクセルの値は (x+y)/2、x*y、x^y など好きな式で構いません。
// テストでは「形 (dy 行 × dx 列)」と「値が (x^y) であること」を確認します。
//
// 実装できたら go run ./pic/cmd で pic.png が生成されます。
//
// 確認: go test ./pic/
package pic

// Pic は dy 行 dx 列の 2 次元スライスを返します。値は x^y (XOR) にしてください。
func Pic(dx, dy int) [][]uint8 {
	// TODO: make で外側のスライスを作り、各行も make で作ってください
	return nil
}
