package system1go

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNoHTTPOnUnsupported(t *testing.T) {
	var calls atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer s.Close()
	c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: s.URL})
	_, err := c.Decide(context.Background(), Request{Input: "x", Questions: []Question{{Name: "typed", Type: Choice, Choices: []Option{{Value: true}, {Value: "true"}}}}})
	var e *Error
	if !errors.As(err, &e) || e.Question != "typed" || e.Feature != "booleanChoices" || e.Protocol != SystemOne || calls.Load() != 0 {
		t.Fatalf("%+v calls=%d", err, calls.Load())
	}
}
func TestConcurrentClientIsolation(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer a" {
			t.Errorf("header %q", r.Header.Get("Authorization"))
		}
		fmt.Fprint(w, `{"answers":{"p":{"type":"noul","noul":1}}}`)
	}))
	defer s.Close()
	h := http.Header{"Authorization": []string{"Bearer a"}}
	c, e := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: s.URL, Headers: h})
	if e != nil {
		t.Fatal(e)
	}
	h.Set("Authorization", "Bearer changed")
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, e := c.Decide(context.Background(), Request{Input: "x", Questions: []Question{{Name: "p", Type: Predicate}}}); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
}
