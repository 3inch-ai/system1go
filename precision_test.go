package system1go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestPrecisionAndHeaderIsolation(t *testing.T) {
	c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: "https://example.test"})
	b, e := c.Encode(Request{Input: map[string]any{"n": uint64(9007199254740993)}, Questions: []Question{{Name: "p", Type: Predicate}}})
	if e != nil || !strings.Contains(string(b), "9007199254740993") {
		t.Fatalf("precision lost %s %v", b, e)
	}
	_, e = NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: "https://example.test", APIKey: "secret", Headers: http.Header{"authorization": []string{"Bearer other"}}})
	if e == nil {
		t.Fatal("case-insensitive auth conflict accepted")
	}
}
func TestScoreAnomalies(t *testing.T) {
	for _, p := range []Protocol{SystemOne, OpenAIDecisions} {
		c, _ := NewClient(Config{Protocol: p, Model: "m", BaseURL: "https://example.test"})
		q := []Question{{Name: "s", Type: Score, Instructions: "x", Levels: []Level{{Label: "a"}, {Label: "b"}}}}
		var a map[string]any
		if p == SystemOne {
			json.Unmarshal([]byte(`{"type":"score","score":0.8,"confidence":0.4,"legend":{"0":"a","1":"b"},"probabilities":{"0":0.2,"1":0.8}}`), &a)
		} else {
			json.Unmarshal([]byte(`{"type":"score","name":"s","score":0.8,"confidence":0.4,"probabilities":[{"value":0,"label":"a","probability":0.2},{"value":1,"label":"b","probability":0.8}]}`), &a)
		}
		for _, field := range []string{"score", "confidence", "probabilities"} {
			saved := a[field]
			for _, v := range []any{nil, "wrong", -1, 2} {
				a[field] = v
				var env any = map[string]any{"answers": map[string]any{"s": a}}
				if p == OpenAIDecisions {
					env = map[string]any{"answers": []any{a}}
				}
				raw, _ := json.Marshal(env)
				d, e := c.Decode(raw, q)
				if e != nil {
					t.Fatal(e)
				}
				if d.Answers["s"].Status != Invalid {
					t.Errorf("accepted %s", raw)
				}
			}
			a[field] = saved
		}
	}
}
func TestSafeErrorDiagnostics(t *testing.T) {
	c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: "https://example.test", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, fmt.Errorf("credential-secret") })}})
	_, err := c.Decide(context.Background(), Request{Input: "x", Questions: []Question{{Name: "p", Type: Predicate}}})
	var e *Error
	if !errors.As(err, &e) || e.Cause() == nil || strings.Contains(fmt.Sprintf("%+v", err), "secret") {
		t.Fatal(err)
	}
	e.raw = []byte("sensitive")
	b := e.Diagnostics()
	b[0] = 'x'
	if string(e.Diagnostics()) != "sensitive" {
		t.Fatal("diagnostic alias")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
