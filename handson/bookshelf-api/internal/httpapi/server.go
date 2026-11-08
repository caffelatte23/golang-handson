// Package httpapi は HTTP ハンドラとルーティングを提供します。
// Go 1.22 以降の標準 net/http だけで、メソッド・パスパラメータ付きのルーティングが書けます。
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/caffelatte23/golang-hanson/handson/bookshelf-api/internal/book"
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

	// ミドルウェアは func(http.Handler) http.Handler で表現し、外側から順に適用する
	var h http.Handler = mux
	h = logging(s.logger)(h)
	h = recoverer(s.logger)(h)
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
