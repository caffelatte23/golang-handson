package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestChecker_Check(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/missing", http.NotFound)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Checker{Client: srv.Client(), Concurrency: 2}
	urls := []string{srv.URL + "/ok", srv.URL + "/missing", "http://127.0.0.1:1/unreachable"}

	got, err := c.Check(context.Background(), urls)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d", len(got))
	}
	// 入力順が保たれていること
	for i := range urls {
		if got[i].URL != urls[i] {
			t.Errorf("got[%d].URL = %s, want %s", i, got[i].URL, urls[i])
		}
	}
	if !got[0].OK() {
		t.Errorf("/ok should be OK: %+v", got[0])
	}
	if got[1].OK() || got[1].Status != http.StatusNotFound {
		t.Errorf("/missing should be 404: %+v", got[1])
	}
	if got[2].Err == nil {
		t.Errorf("unreachable should have error")
	}
}

func TestChecker_ConcurrencyLimit(t *testing.T) {
	var cur, peak atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := cur.Add(1)
		defer cur.Add(-1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}))
	defer srv.Close()

	urls := make([]string, 20)
	for i := range urls {
		urls[i] = srv.URL
	}
	c := &Checker{Client: srv.Client(), Concurrency: 3}
	if _, err := c.Check(context.Background(), urls); err != nil {
		t.Fatal(err)
	}
	if p := peak.Load(); p > 3 {
		t.Errorf("peak concurrency = %d, want <= 3", p)
	}
}

func TestChecker_Cancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	c := &Checker{Client: srv.Client(), Concurrency: 4}
	start := time.Now()
	_, err := c.Check(ctx, []string{srv.URL, srv.URL, srv.URL})
	if err == nil {
		t.Fatal("want error on timeout")
	}
	if time.Since(start) > time.Second {
		t.Errorf("cancellation took too long: %v", time.Since(start))
	}
}
