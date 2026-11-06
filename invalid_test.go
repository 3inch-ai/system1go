package system1go

import (
	"strings"
	"testing"
)

func TestInvalidAnswers(t *testing.T) {
	q := []Question{{Name: "p", Type: Predicate, Instructions: "x"}, {Name: "other", Type: Predicate, Instructions: "x"}}
	c, _ := NewClient(Config{Protocol: OpenAIDecisions, Model: "m", BaseURL: "https://example.test"})
	d, e := c.Decode([]byte(`{"answers":[{"name":"p","type":"refusal"},{"name":"other","type":"predicate","probability":0}]}`), q)
	if e != nil {
		t.Fatal(e)
	}
	if d.Answers["p"].Status != Refused || d.Answers["other"].Status != OK {
		t.Fatalf("%+v", d)
	}
	for _, s := range []string{`{"answers":[]}`, `{"answers":[{"name":"wrong","type":"predicate","probability":0.8}]}`, `{"answers":[{"name":"p","type":"predicate","probability":null}]}`, `{"answers":[{"name":"p","type":"predicate"}]}`, `{"answers":[{"name":"p","type":"predicate","probability":2}]}`, `{"answers":[{"name":"p","type":"predicate","probability":0.8},{"name":"p","type":"predicate","probability":0.8}]}`} {
		d, e := c.Decode([]byte(s), q)
		if e != nil {
			t.Fatal(e)
		}
		if d.Answers["p"].Status != Invalid || len(d.Answers["p"].Issues) == 0 {
			t.Errorf("not invalid with issue: %s %+v", s, d)
		}
	}
	if _, e := c.Decode([]byte(`{"answers":[]} {}`), q); e == nil {
		t.Fatal("trailing JSON accepted")
	}
}
func TestNativeChoiceAndExtensions(t *testing.T) {
	c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: "http://localhost", ExtraBodyKeys: []string{"attention_mode"}})
	r := Request{Input: []any{true}, Questions: []Question{{Name: "c", Type: Choice, Choices: []Option{{Value: "only"}}}}, Options: map[string]any{"attention_mode": "causal"}}
	b, e := c.Encode(r)
	if e != nil || !strings.Contains(string(b), `"attention_mode":"causal"`) {
		t.Fatalf("%s %v", b, e)
	}
	d, e := c.Decode([]byte(`{"answers":{"c":{"type":"choice","choice":"only","confidence":1,"probabilities":{"only":1}}}}`), r.Questions)
	if e != nil || d.Answers["c"].Status != OK {
		t.Fatalf("%+v %v", d, e)
	}
}
