package system1go

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

type countingInput struct{ n *atomic.Int32 }

func (v countingInput) MarshalJSON() ([]byte, error) {
	v.n.Add(1)
	return []byte(`{"text":"evidence"}`), nil
}
func TestSnapshot(t *testing.T) {
	var n atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"answers":{"p":{"type":"noul","noul":0}}}`)
	}))
	defer s.Close()
	c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: s.URL})
	r := Request{Input: countingInput{&n}, Questions: []Question{{Name: "p", Type: Predicate}}}
	if _, e := c.Decide(context.Background(), r); e != nil {
		t.Fatal(e)
	}
	if n.Load() != 1 {
		t.Fatalf("MarshalJSON called %d times", n.Load())
	}
	cyclic := map[string]any{}
	cyclic["self"] = cyclic
	for _, v := range []any{cyclic, math.NaN(), json.RawMessage(`1`)} {
		r.Input = v
		if _, e := c.Encode(r); e == nil {
			t.Fatal("invalid input accepted")
		}
	}
}
func TestMalformedDistributions(t *testing.T) {
	c, _ := NewClient(Config{Protocol: OpenAIDecisions, Model: "m", BaseURL: "https://example.test"})
	q := []Question{{Name: "c", Type: Choice, Instructions: "x", Choices: []Option{{Value: true}, {Value: "true"}}}}
	for _, raw := range []string{
		`{"answers":[{"name":"c","type":"choice","choice":true,"confidence":0.2,"probabilities":[{"value":{},"probability":0.7},{"value":"true","probability":0.3}]}]}`,
		`{"answers":[{"name":"c","type":"choice","choice":true,"confidence":0.2,"probabilities":[{"value":true,"probability":0.1},{"value":"true","probability":0.9}]}]}`,
		`{"answers":[{"name":"c","type":"choice","choice":true,"confidence":0.2,"probabilities":[{"value":true,"probability":0.8},{"value":true,"probability":0.2}]}]}`,
	} {
		d, e := c.Decode([]byte(raw), q)
		if e != nil {
			t.Fatal(e)
		}
		if d.Answers["c"].Status != Invalid {
			t.Errorf("accepted %s", raw)
		}
	}
}
