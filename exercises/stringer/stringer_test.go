package stringer

import (
	"fmt"
	"testing"
)

func TestIPAddr(t *testing.T) {
	hosts := map[string]IPAddr{
		"127.0.0.1": {127, 0, 0, 1},
		"8.8.8.8":   {8, 8, 8, 8},
	}
	for want, ip := range hosts {
		// IPAddr を any として扱い、fmt.Stringer を満たすか確認する
		if _, ok := any(ip).(fmt.Stringer); !ok {
			t.Fatal("IPAddr does not implement fmt.Stringer")
		}
		if got := fmt.Sprint(ip); got != want {
			t.Errorf("fmt.Sprint(%v) = %q, want %q", [4]byte(ip), got, want)
		}
	}
}
