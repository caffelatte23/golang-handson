//go:build !solution

// Package image は第5章の演習です。
//
// image.Image インターフェースを実装する Image 型を作ってください。
//
//	ColorModel() color.Model  → color.RGBAModel
//	Bounds() image.Rectangle  → image.Rect(0, 0, W, H)
//	At(x, y int) color.Color  → color.RGBA{uint8(x ^ y), uint8(x ^ y), 255, 255}
//
// 確認: go test ./image/
package image

// Image は W×H の画像です。
type Image struct {
	W, H int
}

// TODO: ColorModel, Bounds, At メソッドを追加してください
