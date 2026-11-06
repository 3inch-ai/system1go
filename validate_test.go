package system1go

import (
	"math"
	"strings"
	"testing"
)

func TestValidation(t *testing.T) {
	for _, p := range []Protocol{SystemOne, OpenAIDecisions} {
		c, _ := NewClient(Config{Protocol: p, Model: "m", BaseURL: "https://example.test"})
		base := Question{Name: "p", Type: Predicate, Instructions: "x"}
		cases := []Request{{Input: "x"}, {Input: 1, Questions: []Question{base}}, {Input: math.NaN(), Questions: []Question{base}}, {Input: "x", Questions: []Question{base, base}}, {Input: "x", Questions: []Question{{Name: "x", Type: "box"}}}, {Input: "x", Questions: []Question{{Name: "x", Type: Predicate, Instructions: "x", Choices: []Option{{Value: "ignored"}}}}}, {Input: "x", Questions: []Question{{Name: "x", Type: Choice, Instructions: "pick", Choices: []Option{{Value: []string{"bad"}}, {Value: "x"}}}}}, {Input: "x", Questions: []Question{base}, Options: map[string]any{"unknown": "secret"}}}
		for i, r := range cases {
			if _, e := c.Encode(r); e == nil {
				t.Errorf("%s case %d accepted", p, i)
			} else if strings.Contains(e.Error(), "secret") {
				t.Fatal("leak")
			}
		}
	}
}
func TestOptions(t *testing.T) {
	c, _ := NewClient(Config{Protocol: OpenAIDecisions, Model: "m", BaseURL: "https://example.test"})
	r := Request{Input: "x", Questions: []Question{{Name: "p", Type: Predicate, Instructions: "x"}}, Options: map[string]any{"prompt_cache_key": nil, "prompt_cache_options": map[string]any{"mode": "explicit", "ttl": "30m"}, "safety_identifier": "user", "inputTextCacheBreakpoint": true}}
	b, e := c.Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	for _, s := range []string{`"prompt_cache_key":null`, `"safety_identifier":"user"`, `"prompt_cache_breakpoint"`} {
		if !strings.Contains(string(b), s) {
			t.Errorf("missing %s: %s", s, b)
		}
	}
}
