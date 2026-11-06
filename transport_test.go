package system1go

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTransportSafety(t *testing.T) {
	r := Request{Input: "x", Questions: []Question{{Name: "p", Type: Predicate, Instructions: "x"}}}
	for _, status := range []int{401, 403, 429, 400, 422, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(status)
				fmt.Fprint(w, "private secret")
			}))
			defer s.Close()
			c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: s.URL})
			_, err := c.Decide(context.Background(), r)
			var e *Error
			if !errors.As(err, &e) || e.Status != status || strings.Contains(fmt.Sprintf("%+v", err), "secret") || e.HTTP.Headers.Get("Retry-After") != "1" {
				t.Fatalf("%+v", err)
			}
			want := "http"
			if status == 401 || status == 403 {
				want = "authentication"
			}
			if status == 429 {
				want = "rate_limit"
			}
			if status == 400 || status == 422 {
				want = "invalid_request"
			}
			if e.Code != want {
				t.Fatalf("code %s", e.Code)
			}
		})
	}
	t.Run("bounded", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, strings.Repeat("x", 500)) }))
		defer s.Close()
		c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: s.URL, MaxResponseBytes: 64})
		_, e := c.Decide(context.Background(), r)
		var se *Error
		if !errors.As(e, &se) || se.Code != "response_too_large" {
			t.Fatal(e)
		}
	})
	t.Run("redirect", func(t *testing.T) {
		var hits atomic.Int32
		target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1) }))
		defer target.Close()
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 307) }))
		defer s.Close()
		h := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return nil }}
		c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: s.URL, APIKey: "secret", HTTPClient: h})
		if _, e := c.Decide(context.Background(), r); e == nil {
			t.Fatal("redirect accepted")
		}
		if hits.Load() != 0 {
			t.Fatal("credential redirect followed")
		}
	})
	t.Run("timeout-body", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		}))
		defer s.Close()
		c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: s.URL, Timeout: 20 * time.Millisecond})
		_, e := c.Decide(context.Background(), r)
		var se *Error
		if !errors.As(e, &se) || se.Code != "timeout" || !errors.Is(e, context.DeadlineExceeded) {
			t.Fatalf("%+v", e)
		}
	})
	t.Run("cancel", func(t *testing.T) {
		c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: "https://example.test"})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, e := c.Decide(ctx, r)
		var se *Error
		if !errors.As(e, &se) || se.Code != "aborted" || !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
	})
	for _, key := range []string{"secret\nvalue", "secret\rvalue"} {
		_, e := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: "https://example.test", APIKey: key})
		if e == nil || strings.Contains(e.Error(), "secret") {
			t.Fatal("unsafe key validation")
		}
	}
}
