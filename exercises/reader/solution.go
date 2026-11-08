//go:build solution

package reader

// MyReader は 'A' を無限に返す Reader です。
type MyReader struct{}

// Read は b をすべて 'A' で埋めます。
func (MyReader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 'A'
	}
	return len(b), nil
}
