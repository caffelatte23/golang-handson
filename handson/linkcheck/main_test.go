package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("all ok via args", func(t *testing.T) {
		var out, errOut bytes.Buffer
		code := run([]string{srv.URL + "/ok"}, strings.NewReader(""), &out, &errOut)
		if code != 0 {
			t.Fatalf("code = %d, out=%s err=%s", code, &out, &errOut)
		}
	})
	t.Run("broken via stdin", func(t *testing.T) {
		var out, errOut bytes.Buffer
		in := strings.NewReader("# comment\n" + srv.URL + "/ok\n" + srv.URL + "/nope\n")
		code := run(nil, in, &out, &errOut)
		if code != 1 || !strings.Contains(out.String(), "2 checked, 1 broken") {
			t.Fatalf("code = %d, out=%s", code, &out)
		}
	})
	t.Run("no input", func(t *testing.T) {
		var out, errOut bytes.Buffer
		if code := run(nil, strings.NewReader(""), &out, &errOut); code != 2 {
			t.Fatalf("code = %d", code)
		}
	})
}
