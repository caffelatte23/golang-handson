// Pic の結果を PNG 画像 (pic.png) として保存します。
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/caffelatte23/golang-hanson/exercises/pic"
)

func main() {
	const dx, dy = 256, 256
	data := pic.Pic(dx, dy)
	img := image.NewGray(image.Rect(0, 0, dx, dy))
	for y := range data {
		for x := range data[y] {
			img.SetGray(x, y, color.Gray{Y: data[y][x]})
		}
	}
	f, err := os.Create("pic.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil { // 書き込みのエラーは Close で報告されることもある
		log.Fatal(err)
	}
	log.Println("wrote pic.png")
}
