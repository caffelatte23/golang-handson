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
