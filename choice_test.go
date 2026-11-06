package system1go

import (
	"encoding/json"
	"testing"
)

func TestTypedChoice(t *testing.T) {
	c, e := NewClient(Config{Protocol: OpenAIDecisions, Model: "gpt-6-luna", BaseURL: "https://example.test"})
	if e != nil {
		t.Fatal(e)
	}
	r := Request{Input: map[string]any{"text": "x"}, Questions: []Question{{Name: "c", Type: Choice, Instructions: "pick", Choices: []Option{{Value: true}, {Value: "true"}}}}}
	b, e := c.Encode(r)
	if e != nil {
		t.Fatal(e)
	}
	var wire map[string]any
	json.Unmarshal(b, &wire)
	if _, ok := wire["input"].([]any); !ok {
		t.Fatalf("JSON input not native: %s", b)
	}
	d, e := c.Decode([]byte(`{"answers":[{"type":"choice","name":"c","choice":true,"confidence":0.6,"probabilities":[{"value":true,"probability":0.7},{"value":"true","probability":0.3}]}]}`), r.Questions)
	if e != nil {
		t.Fatal(e)
	}
	a := d.Answers["c"]
	if a.Status != OK || a.Choice != true || len(a.Probabilities) != 2 || a.Probabilities[1].Value != "true" {
		t.Fatalf("%+v", a)
	}
}
