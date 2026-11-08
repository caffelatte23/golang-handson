//go:build !solution

// Package rot13 は第5章の演習です。
//
// 別の io.Reader をラップし、読み出したアルファベットを ROT13 (13 文字ずらす)
// 変換して返す Rot13Reader を実装してください。アルファベット以外はそのまま返します。
//
//	s := strings.NewReader("Lbh penpxrq gur pbqr!")
//	io.Copy(os.Stdout, Rot13Reader{R: s}) // → You cracked the code!
//
// 確認: go test ./rot13/
package rot13

import "io"

// Rot13Reader は R から読んだ内容を ROT13 変換します。
type Rot13Reader struct {
	R io.Reader
}

// TODO: func (r Rot13Reader) Read(b []byte) (int, error) を追加してください
