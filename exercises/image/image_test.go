package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestImage(t *testing.T) {
	m, ok := any(Image{W: 16, H: 8}).(image.Image)
	if !ok {
		t.Fatal("Image does not implement image.Image")
	}
	if m.ColorModel() != color.RGBAModel {
		t.Errorf("ColorModel = %v, want color.RGBAModel", m.ColorModel())
	}
	if b := m.Bounds(); b != image.Rect(0, 0, 16, 8) {
		t.Errorf("Bounds = %v", b)
	}
	if got, want := m.At(3, 5), (color.RGBA{6, 6, 255, 255}); got != want {
		t.Errorf("At(3,5) = %v, want %v", got, want)
	}
	// image.Image を満たせば、標準の PNG エンコーダにそのまま渡せる
	var buf bytes.Buffer
	if err := png.Encode(&buf, m); err != nil {
		t.Fatal(err)
	}
}
