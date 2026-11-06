package system1go

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestScoreIndexPresence(t *testing.T) {
	c, _ := NewClient(Config{Protocol: OpenAIDecisions, Model: "m", BaseURL: "https://example.test"})
	q := []Question{{Name: "s", Type: Score, Instructions: "x", Levels: []Level{{Label: "a"}, {Label: "b"}}}}
	for _, field := range []string{"", `"value":null,`, `"value":0.5,`} {
		raw := `{"answers":[{"name":"s","type":"score","score":0.8,"confidence":0.4,"probabilities":[{` + field + `"label":"a","probability":0.2},{"value":1,"label":"b","probability":0.8}]}]}`
		d, e := c.Decode([]byte(raw), q)
		if e != nil {
			t.Fatal(e)
		}
		if d.Answers["s"].Status != Invalid {
			t.Errorf("missing/null/fractional index accepted: %s", raw)
		}
	}
}
func TestScopeLimits(t *testing.T) {
	for _, p := range []Protocol{SystemOne, OpenAIDecisions} {
		c, _ := NewClient(Config{Protocol: p, Model: "m", BaseURL: "https://example.test"})
		q := Question{Name: "p", Type: Predicate, Instructions: "x"}
		r := Request{Input: "x", Questions: []Question{q}}
		if p == OpenAIDecisions {
			r.Questions = make([]Question, 401)
			for i := range r.Questions {
				r.Questions[i] = q
			}
			if _, e := c.Encode(r); e == nil {
				t.Fatal("question bound")
			}
			r.Questions = []Question{q}
			r.Input = []any{true}
			if _, e := c.Encode(r); e == nil {
				t.Fatal("non-object JSON array")
			}
			r.Input = map[string]any{"x": strings.Repeat("a", 262145)}
			if _, e := c.Encode(r); e == nil {
				t.Fatal("JSON byte bound")
			}
			r.Input = "x"
			r.Options = map[string]any{"prompt_cache_options": map[string]any{"mode": "explicit"}}
			if _, e := c.Encode(r); e != nil {
				t.Fatal("explicit no marker is valid", e)
			}
		}
	}
	c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: "https://example.test"})
	for _, raw := range []string{`null`, `[]`, `{}`, `{"answers":null}`, `{"answers":[]}`} {
		if _, e := c.Decode([]byte(raw), []Question{{Name: "p", Type: Predicate}}); e == nil {
			t.Fatal("envelope", raw)
		}
	}
	r := Request{Input: json.RawMessage(`{"text":"x"}`), Questions: []Question{{Name: "p", Type: Predicate}}}
	if _, e := c.Encode(r); e != nil {
		t.Fatal(e)
	}
}
