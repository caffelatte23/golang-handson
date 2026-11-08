//go:build solution

package image

import (
	"image"
	"image/color"
)

// Image は W×H の画像です。
type Image struct {
	W, H int
}

// ColorModel は色モデルを返します。
func (Image) ColorModel() color.Model { return color.RGBAModel }

// Bounds は画像の範囲を返します。
func (m Image) Bounds() image.Rectangle { return image.Rect(0, 0, m.W, m.H) }

// At は (x, y) の色を返します。
func (Image) At(x, y int) color.Color {
	v := uint8(x ^ y)
	return color.RGBA{R: v, G: v, B: 255, A: 255}
}
