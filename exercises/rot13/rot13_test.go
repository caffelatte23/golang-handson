package rot13

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestRot13Reader(t *testing.T) {
	r := asReader(t, Rot13Reader{R: strings.NewReader("Lbh penpxrq gur pbqr! 123")})
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if want := "You cracked the code! 123"; string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRot13Reader_SmallReads(t *testing.T) {
	// 1 バイトずつしか返さない Reader でも正しく動くか (iotest は標準のテスト補助)
	r := asReader(t, Rot13Reader{R: iotest.OneByteReader(strings.NewReader("Uryyb, Tbcure"))})
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if want := "Hello, Gopher"; string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// asReader は v が io.Reader を実装していれば返し、していなければテストを失敗させます。
// (未実装の状態でもテストがコンパイルできるように any 経由で確認しています)
func asReader(t *testing.T, v any) io.Reader {
	t.Helper()
	r, ok := v.(io.Reader)
	if !ok {
		t.Fatalf("%T does not implement io.Reader (Read メソッドが必要です)", v)
	}
	return r
}
