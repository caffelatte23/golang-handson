//go:build solution

package rot13

import "io"

// Rot13Reader は R から読んだ内容を ROT13 変換します。
type Rot13Reader struct {
	R io.Reader
}

// Read は内側の Reader から読み、読めた分だけ変換します。
func (r Rot13Reader) Read(b []byte) (int, error) {
	n, err := r.R.Read(b)
	for i := range b[:n] { // 読めたバイト数 n までだけを変換する
		b[i] = rot13(b[i])
	}
	return n, err
}

func rot13(c byte) byte {
	switch {
	case 'a' <= c && c <= 'z':
		return 'a' + (c-'a'+13)%26
	case 'A' <= c && c <= 'Z':
		return 'A' + (c-'A'+13)%26
	}
	return c
}
