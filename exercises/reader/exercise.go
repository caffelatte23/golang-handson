//go:build !solution

// Package reader は第5章の演習です。
//
// 'A' の文字を無限に返し続ける io.Reader を実装してください。
// Read(b []byte) は b を 'A' で埋め、len(b) と nil を返します。
//
// 確認: go test ./reader/
package reader

// MyReader は 'A' を無限に返す Reader です。
type MyReader struct{}

// TODO: func (MyReader) Read(b []byte) (int, error) を追加してください
