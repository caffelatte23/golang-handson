//go:build !solution

// Package stringer は第5章の演習です。
//
// IPAddr 型に fmt.Stringer インターフェースを実装し、
// IPAddr{1, 2, 3, 4} が "1.2.3.4" と表示されるようにしてください。
//
// 確認: go test ./stringer/
package stringer

// IPAddr は IPv4 アドレスです。
type IPAddr [4]byte

// TODO: func (ip IPAddr) String() string を追加してください
