//go:build solution

package stringer

import "fmt"

// IPAddr は IPv4 アドレスです。
type IPAddr [4]byte

// String は "1.2.3.4" 形式の文字列を返します。
func (ip IPAddr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}
