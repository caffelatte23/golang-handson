package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caffelatte23/golang-hanson/handson/bookshelf-api/internal/book"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httptest.NewServer(New(book.NewMemoryRepository(), logger).Handler())
	t.Cleanup(srv.Close)
	return srv
}

// response はテストで使う最小限のレスポンス情報です。
// *http.Response を返すと Body の Close 責任が曖昧になるため、ここで読み切ります。
type response struct {
	StatusCode int
	Header     http.Header
}

func do(t *testing.T, method, url, body string) (response, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return response{StatusCode: res.StatusCode, Header: res.Header}, string(b)
}

func TestBooksCRUD(t *testing.T) {
	srv := newTestServer(t)

	res, body := do(t, http.MethodPost, srv.URL+"/books", `{"title":"プログラミング言語Go","author":"Donovan & Kernighan"}`)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: status=%d body=%s", res.StatusCode, body)
	}
	if loc := res.Header.Get("Location"); loc != "/books/1" {
		t.Errorf("Location = %q", loc)
	}
	var created book.Book
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatal(err)
	}

	res, body = do(t, http.MethodGet, srv.URL+"/books/1", "")
	if res.StatusCode != http.StatusOK || !strings.Contains(body, "プログラミング言語Go") {
		t.Fatalf("get: status=%d body=%s", res.StatusCode, body)
	}

	res, body = do(t, http.MethodPut, srv.URL+"/books/1", `{"title":"プログラミング言語Go","author":"Donovan & Kernighan","read":true}`)
	if res.StatusCode != http.StatusOK || !strings.Contains(body, `"read":true`) {
		t.Fatalf("update: status=%d body=%s", res.StatusCode, body)
	}

	res, _ = do(t, http.MethodDelete, srv.URL+"/books/1", "")
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: status=%d", res.StatusCode)
	}

	res, _ = do(t, http.MethodGet, srv.URL+"/books/1", "")
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete: status=%d", res.StatusCode)
	}
}

func TestErrors(t *testing.T) {
	srv := newTestServer(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"不正なJSON", http.MethodPost, "/books", `{`, http.StatusBadRequest},
		{"未知のフィールド", http.MethodPost, "/books", `{"title":"a","author":"b","x":1}`, http.StatusBadRequest},
		{"バリデーションエラー", http.MethodPost, "/books", `{"title":"","author":""}`, http.StatusUnprocessableEntity},
		{"不正なID", http.MethodGet, "/books/abc", "", http.StatusBadRequest},
		{"存在しないID", http.MethodGet, "/books/999", "", http.StatusNotFound},
		{"許可されないメソッド", http.MethodPatch, "/books/1", "", http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, body := do(t, tt.method, srv.URL+tt.path, tt.body)
			if res.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d (body=%s)", res.StatusCode, tt.wantStatus, body)
			}
		})
	}
}
