# 第15章 ハンズオン② Web API — 書籍管理 REST API (net/http + slog)

Go はクラウドネイティブなバックエンド開発で広く使われています。この章では **外部フレームワークを使わず、標準ライブラリだけで** 本番運用を意識した REST API を作ります。

📁 完成コード: [`handson/bookshelf-api/`](../handson/bookshelf-api/)

## 学べること

- Go 1.22+ の `http.ServeMux` によるメソッド・パスパラメータ付きルーティング
- JSON のエンコード/デコードと入力バリデーション
- ドメインエラー → HTTP ステータスへの変換 (`errors.Is` / `errors.As`)
- ミドルウェア (アクセスログ・panic リカバリ) の書き方
- `log/slog` による構造化ログ
- `sync.RWMutex` による並行安全なリポジトリ
- タイムアウト設定とグレースフルシャットダウン
- `httptest` を使った API テスト

## API 仕様

| メソッド | パス | 説明 | 成功時 |
| --- | --- | --- | --- |
| GET | `/healthz` | ヘルスチェック | 200 |
| GET | `/books` | 一覧 | 200 |
| POST | `/books` | 作成 | 201 + `Location` ヘッダ |
| GET | `/books/{id}` | 取得 | 200 |
| PUT | `/books/{id}` | 更新 | 200 |
| DELETE | `/books/{id}` | 削除 | 204 |

アプリケーションのエラーは `{"error": "...", "details": [...]}` 形式の JSON で返します (400 / 404 / 422 / 500)。ただし、**存在しないパスへの 404 と、メソッドが合わないときの 405 は `ServeMux` が返すので、プレーンテキスト** (`404 page not found` など) になります。すべてを JSON にそろえたい場合は、発展課題を参照してください。

## Step 1. プロジェクトを作る

```bash
mkdir bookshelf-api && cd bookshelf-api
go mod init github.com/<your-name>/bookshelf-api
mkdir -p cmd/server internal/book internal/httpapi
```

```
bookshelf-api/
├── cmd/server/main.go       ← 組み立てと起動 (DI・設定・シャットダウン)
├── internal/
│   ├── book/                ← ドメイン層: HTTP を知らない
│   │   ├── book.go          ← エンティティ・バリデーション・Repository interface
│   │   ├── memory.go        ← インメモリ実装
│   │   └── book_test.go
│   └── httpapi/             ← HTTP 層
│       ├── server.go        ← ルーティング・ハンドラ
│       ├── middleware.go
│       └── server_test.go
├── Dockerfile
└── go.mod
```

第10章で見た公式レイアウト (c) と同じ構成です。**外部依存はゼロ** です。

## Step 2. ドメインモデルと Repository インターフェース

`internal/book/book.go`:

```go
// handson/bookshelf-api/internal/book/book.go
// Package book は書籍のドメインモデルとリポジトリを定義します。
package book

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("book not found")
	ErrValidation = errors.New("validation error")
)

// Book は書籍を表すエンティティです。
type Book struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Input は作成・更新時の入力です。
type Input struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Read   bool   `json:"read"`
}

// Validate は入力を検証します。エラーは ErrValidation をラップします。
func (in *Input) Validate() error {
	in.Title = strings.TrimSpace(in.Title)
	in.Author = strings.TrimSpace(in.Author)
	var problems []string
	if in.Title == "" {
		problems = append(problems, "title is required")
	}
	if len([]rune(in.Title)) > 200 {
		problems = append(problems, "title must be <= 200 characters")
	}
	if in.Author == "" {
		problems = append(problems, "author is required")
	}
	if len(problems) > 0 {
		return &ValidationError{Problems: problems}
	}
	return nil
}

// ValidationError は複数の検証エラーを保持するカスタムエラー型です。
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return "validation error: " + strings.Join(e.Problems, ", ")
}

// Is を実装すると errors.Is(err, ErrValidation) が true になります。
func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

// Repository は永続化層の抽象です。
// 実装はインメモリ (memory.go) ですが、PostgreSQL などに差し替えられます。
type Repository interface {
	List(ctx context.Context) ([]Book, error)
	Get(ctx context.Context, id int64) (Book, error)
	Create(ctx context.Context, in Input) (Book, error)
	Update(ctx context.Context, id int64, in Input) (Book, error)
	Delete(ctx context.Context, id int64) error
}
```

**ポイント**

- `ValidationError` は **独自エラー型** で、問題点のリストを持ちます。`Is` メソッドを実装しているので `errors.Is(err, ErrValidation)` でも判定できます。
- `Repository` の各メソッドは第 1 引数に `context.Context` を取ります。DB 実装に差し替えたとき、リクエストのキャンセルをクエリまで伝播させるためです。

## Step 3. インメモリのリポジトリ

`internal/book/memory.go`:

```go
// handson/bookshelf-api/internal/book/memory.go
package book

import (
	"cmp"
	"context"
	"slices"
	"sync"
	"time"
)

// MemoryRepository はスレッドセーフなインメモリ実装です。
// net/http のハンドラは並行に呼ばれるため、mutex で保護します。
type MemoryRepository struct {
	mu     sync.RWMutex
	books  map[int64]Book
	nextID int64
	now    func() time.Time
}

// コンパイル時にインターフェースを満たしているか確認するイディオム。
var _ Repository = (*MemoryRepository)(nil)

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{books: make(map[int64]Book), nextID: 1, now: time.Now}
}

func (r *MemoryRepository) List(ctx context.Context) ([]Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Book, 0, len(r.books))
	for _, b := range r.books {
		out = append(out, b)
	}
	// map の反復順は不定なので ID でソートして返す
	slices.SortFunc(out, func(a, b Book) int { return cmp.Compare(a.ID, b.ID) })
	return out, nil
}

func (r *MemoryRepository) Get(ctx context.Context, id int64) (Book, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.books[id]
	if !ok {
		return Book{}, ErrNotFound
	}
	return b, nil
}

func (r *MemoryRepository) Create(ctx context.Context, in Input) (Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	b := Book{ID: r.nextID, Title: in.Title, Author: in.Author, Read: in.Read, CreatedAt: now, UpdatedAt: now}
	r.books[b.ID] = b
	r.nextID++
	return b, nil
}

func (r *MemoryRepository) Update(ctx context.Context, id int64, in Input) (Book, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.books[id]
	if !ok {
		return Book{}, ErrNotFound
	}
	b.Title, b.Author, b.Read, b.UpdatedAt = in.Title, in.Author, in.Read, r.now()
	r.books[id] = b
	return b, nil
}

func (r *MemoryRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.books[id]; !ok {
		return ErrNotFound
	}
	delete(r.books, id)
	return nil
}
```

**ポイント**

- **HTTP ハンドラは goroutine ごとに並行実行される** ため、共有している map は必ず mutex で保護します。読み取りが多いので `RWMutex` を使い、`List`/`Get` は `RLock` にしています。
- `var _ Repository = (*MemoryRepository)(nil)` でインターフェースの実装漏れをコンパイル時に検出します。
- map の反復順はランダムなので、`slices.SortFunc` で ID 順に並べています。

並行安全性はテストで確認しましょう (`book_test.go` の `TestMemoryRepository_Concurrent`)。

```bash
go test -race ./internal/book/
```

## Step 4. ハンドラとルーティング

`internal/httpapi/server.go`:

```go
// handson/bookshelf-api/internal/httpapi/server.go
// Package httpapi は HTTP ハンドラとルーティングを提供します。
// Go 1.22 以降の標準 net/http だけで、メソッド・パスパラメータ付きのルーティングが書けます。
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/caffelatte23/golang-handson/handson/bookshelf-api/internal/book"
)

// Server は依存 (Repository, Logger) を保持し、http.Handler を組み立てます。
type Server struct {
	repo   book.Repository
	logger *slog.Logger
}

func New(repo book.Repository, logger *slog.Logger) *Server {
	return &Server{repo: repo, logger: logger}
}

// Handler はミドルウェアを適用したルーターを返します。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /books", s.listBooks)
	mux.HandleFunc("POST /books", s.createBook)
	mux.HandleFunc("GET /books/{id}", s.getBook)
	mux.HandleFunc("PUT /books/{id}", s.updateBook)
	mux.HandleFunc("DELETE /books/{id}", s.deleteBook)

	// ミドルウェアは func(http.Handler) http.Handler で表現する。
	// 後から適用したものほど外側になる: logging(recoverer(mux))。
	// logging を外側に置くことで、panic して recoverer が 500 を返したリクエストもログに残る
	var h http.Handler = mux
	h = recoverer(s.logger)(h)
	h = logging(s.logger)(h)
	return h
}

func (s *Server) listBooks(w http.ResponseWriter, r *http.Request) {
	books, err := s.repo.List(r.Context())
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (s *Server) getBook(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	b, err := s.repo.Get(r.Context(), id)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) createBook(w http.ResponseWriter, r *http.Request) {
	in, ok := s.decodeInput(w, r)
	if !ok {
		return
	}
	b, err := s.repo.Create(r.Context(), in)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	w.Header().Set("Location", "/books/"+strconv.FormatInt(b.ID, 10))
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) updateBook(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	in, ok := s.decodeInput(w, r)
	if !ok {
		return
	}
	b, err := s.repo.Update(r.Context(), id, in)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) deleteBook(w http.ResponseWriter, r *http.Request) {
	id, ok := s.pathID(w, r)
	if !ok {
		return
	}
	if err := s.repo.Delete(r.Context(), id); err != nil {
		s.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---------------------------------------------------------------

func (s *Server) pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64) // Go 1.22+
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid id"})
		return 0, false
	}
	return id, true
}

func (s *Server) decodeInput(w http.ResponseWriter, r *http.Request) (book.Input, bool) {
	var in book.Input
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MiB 制限
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON: " + err.Error()})
		return in, false
	}
	if err := in.Validate(); err != nil {
		s.writeError(w, r, err)
		return in, false
	}
	return in, true
}

type errorBody struct {
	Error   string   `json:"error"`
	Details []string `json:"details,omitempty"`
}

// writeError はドメインのエラーを HTTP ステータスに変換します。
// errors.Is / errors.As による判定は Go のエラー処理の基本パターンです。
func (s *Server) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var verr *book.ValidationError
	switch {
	case errors.As(err, &verr):
		writeJSON(w, http.StatusUnprocessableEntity, errorBody{Error: "validation failed", Details: verr.Problems})
	case errors.Is(err, book.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
	default:
		s.logger.ErrorContext(r.Context(), "internal error", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: "internal server error"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
```

**ポイント**

- `mux.HandleFunc("GET /books/{id}", ...)` と `r.PathValue("id")` は **Go 1.22 の新機能**。これ以前は chi や gorilla/mux が必要でした。メソッドが合わないと自動で `405 Method Not Allowed` (プレーンテキスト) を返します。
- `json.Decoder.DisallowUnknownFields()` でタイプミスしたフィールドを検出し、`http.MaxBytesReader` で巨大なリクエストを防ぎます。
- `writeError` で **ドメインエラー → HTTP ステータス** の変換を 1 か所に集約。ハンドラはエラーをそのまま渡すだけです。
- 500 エラーの詳細はログにだけ出し、クライアントには返しません (情報漏えい防止)。

## Step 5. ミドルウェア

`internal/httpapi/middleware.go`:

```go
// handson/bookshelf-api/internal/httpapi/middleware.go
package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// statusRecorder はレスポンスのステータスコードを記録するためのラッパーです。
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Unwrap は元の ResponseWriter を返します。
// http.ResponseController はこれを使ってラップの内側をたどるので、
// Flush や Hijack などの機能がミドルウェアを通しても使えます (Go 1.20+)。
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// logging はアクセスログを構造化ログ (log/slog) で出力するミドルウェアです。
func logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			logger.InfoContext(r.Context(), "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// recoverer はハンドラ内の panic を捕まえて 500 を返すミドルウェアです。
func recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					if err, ok := v.(error); ok && errors.Is(err, http.ErrAbortHandler) {
						panic(v)
					}
					logger.ErrorContext(r.Context(), "panic recovered", "panic", v, "stack", string(debug.Stack()))
					writeJSON(w, http.StatusInternalServerError, errorBody{Error: "internal server error"})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
```

**ポイント**

- Go のミドルウェアは `func(http.Handler) http.Handler` という **ただの関数** です。chi や他のライブラリのミドルウェアもこの形なので相互運用できます。
- 適用する順番に注意しましょう。`server.go` では `recoverer` → `logging` の順に包んでいるので、外側から `logging(recoverer(mux))` となります。`logging` を外側に置くと、panic して `recoverer` が 500 を返したリクエストもアクセスログに残ります。逆にすると、panic したリクエストはログに残りません。
- `http.ResponseWriter` を埋め込んだ構造体で `WriteHeader` を上書きし、ステータスコードを記録しています (埋め込みによる委譲、第5章)。
- ラップすると、元の ResponseWriter が持っていた `http.Flusher` などのインターフェースが外から見えなくなります。`Unwrap()` メソッドを用意しておくと、`http.ResponseController` が内側をたどって `Flush` などを呼べるようになります。
- `log/slog` の `InfoContext` に `r.Context()` を渡しておくと、後で OpenTelemetry のトレース ID をログに載せる Handler を足すときに役立ちます。

## Step 6. main — 組み立て・設定・グレースフルシャットダウン

`cmd/server/main.go`:

```go
// handson/bookshelf-api/cmd/server/main.go
// server は bookshelf API のエントリポイントです。
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caffelatte23/golang-handson/handson/bookshelf-api/internal/book"
	"github.com/caffelatte23/golang-handson/handson/bookshelf-api/internal/httpapi"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

// run にロジックを寄せ、main は終了コードの決定だけを行うのが定番パターンです。
func run() error {
	addr := flag.String("addr", envOr("ADDR", ":8080"), "listen address")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	repo := book.NewMemoryRepository()
	srv := &http.Server{
		Addr:              *addr,
		Handler:           httpapi.New(repo, logger).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// SIGINT / SIGTERM を受けたら ctx がキャンセルされる
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", *addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	// グレースフルシャットダウン: 処理中のリクエストを最大 10 秒待つ
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
```

**ポイント**

- `main()` は `run()` を呼んで終了コードを決めるだけ。`run()` は `error` を返すので、`defer` が正しく実行されます。
- **`http.Server` のタイムアウトを必ず設定する**。`http.ListenAndServe(addr, h)` の簡易版はタイムアウト無しで、Slowloris 攻撃などに弱くなります (`gosec` の G114 で警告されます)。
- `signal.NotifyContext` で SIGINT/SIGTERM を受けると `ctx` がキャンセルされ、`srv.Shutdown` が **処理中のリクエストの完了を待ってから** 終了します。Kubernetes の Pod 終了時にもこの挙動が重要です。
- 依存 (`repo`, `logger`) は main で生成して手で注入します。DI コンテナは不要です。

## Step 7. 動かしてみる

```bash
go run ./cmd/server
# {"time":"...","level":"INFO","msg":"listening","addr":":8080"}
```

別のターミナルで:

```bash
# 作成
curl -i -X POST localhost:8080/books \
  -d '{"title":"プログラミング言語Go","author":"Donovan & Kernighan"}'
# HTTP/1.1 201 Created
# Location: /books/1
# {"id":1,"title":"プログラミング言語Go",...}

# 一覧・取得・更新・削除
curl localhost:8080/books
curl localhost:8080/books/1
curl -X PUT localhost:8080/books/1 -d '{"title":"プログラミング言語Go","author":"Donovan & Kernighan","read":true}'
curl -i -X DELETE localhost:8080/books/1

# バリデーションエラー
curl -X POST localhost:8080/books -d '{"title":""}'
# {"error":"validation failed","details":["title is required","author is required"]}
```

サーバーのログは JSON 形式で出力されます。

```json
{"time":"...","level":"INFO","msg":"request","method":"POST","path":"/books","status":201,"duration":120287}
```

`Ctrl+C` で `shutting down` と出て正常終了することも確認しましょう。

## Step 8. API テスト

`internal/httpapi/server_test.go`:

```go
// handson/bookshelf-api/internal/httpapi/server_test.go
package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caffelatte23/golang-handson/handson/bookshelf-api/internal/book"
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
```

**ポイント**

- `httptest.NewServer` は実際にローカルポートで HTTP サーバーを起動するので、ルーティングやミドルウェアも含めた **E2E に近いテスト** になります。
- エラーケースはテーブル駆動で網羅しています。
- テスト用ロガーは `io.Discard` に出力して静かにしています。

```bash
go test -race -cover ./...
```

## 📦 OSS と比べてみよう

| 観点 | 本章 | Gitea | Kubernetes API Server |
| --- | --- | --- | --- |
| ルーター | 標準 `ServeMux` | chi | 標準 `net/http` + 独自 (go-restful) |
| ログ | slog | 独自 (`modules/log`) | klog |
| DB | インメモリ | xorm | etcd |
| レイヤー | `book` / `httpapi` | `models` / `services` / `routers` | `pkg/registry` / `apiserver` |

規模が大きくなるほど独自の仕組みを持つようになりますが、**「HTTP 層とドメイン層を分け、依存をインターフェースで受ける」** という基本は同じです。

## 🚀 発展課題

1. **PostgreSQL 対応**: [sqlc](https://sqlc.dev/) + [pgx](https://github.com/jackc/pgx) で `book.Repository` を実装し、`main` で差し替えよう。マイグレーションは [goose](https://github.com/pressly/goose) で。
2. **ページネーション**: `GET /books?limit=20&offset=40` に対応しよう。
3. **認証**: `Authorization: Bearer <token>` を検証するミドルウェアを書こう。
4. **リクエスト ID**: ミドルウェアで ID を生成して `context` に入れ、ログに出力しよう。
5. **メトリクス**: [`prometheus/client_golang`](https://github.com/prometheus/client_golang) で `/metrics` を公開しよう。
6. **エラー形式の統一**: 存在しないパスやメソッド違いのときも JSON でエラーを返すようにしよう (ヒント: `ServeMux` に渡す前にパターンを判定するか、レスポンスをラップしてステータス 404/405 のときに本文を差し替える)。
7. **OpenAPI**: [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) で OpenAPI 定義からハンドラのインターフェースを生成する方式に書き換えよう。
8. **ルーター比較**: 同じ API を chi / gin / echo で書き直し、書き味を比べてみよう。
