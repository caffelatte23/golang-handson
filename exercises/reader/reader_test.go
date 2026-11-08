package reader

import (
	"io"
	"testing"
)

func TestMyReader(t *testing.T) {
	r, ok := any(MyReader{}).(io.Reader)
	if !ok {
		t.Fatal("MyReader does not implement io.Reader")
	}
	for _, size := range []int{1, 7, 1024} {
		b := make([]byte, size)
		n, err := r.Read(b)
		if err != nil || n != size {
			t.Fatalf("Read(%d bytes) = %d, %v", size, n, err)
		}
		for i, c := range b {
			if c != 'A' {
				t.Fatalf("b[%d] = %q, want 'A'", i, c)
			}
		}
	}
}
