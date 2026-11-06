package system1go

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPublicDecodeRejectsInvalidContract(t *testing.T) {
	c, _ := OpenAI(Config{})
	_, e := c.Decode([]byte(`{"answers":[{"name":"c","type":"choice","choice":{},"confidence":1,"probabilities":[{"value":{},"probability":0.7},{"value":"x","probability":0.3}]}]}`), []Question{{Name: "c", Type: Choice, Instructions: "x", Choices: []Option{{Value: map[string]any{}}, {Value: "x"}}}})
	if e == nil {
		t.Fatal("bad contract accepted")
	}
}
func TestHeaderSyntaxBeforeOverwrite(t *testing.T) {
	for _, h := range []http.Header{{"Bad Header": []string{"secret"}}, {"Content-Type": []string{"secret\ninvalid"}}, {"authorization": []string{""}}} {
		_, e := Jev(Config{APIKey: "secret-key", Headers: h})
		if e == nil || strings.Contains(e.Error(), "secret") {
			t.Fatalf("unsafe header config %v", e)
		}
	}
}
func TestEncodedBodyBudget(t *testing.T) {
	c, _ := Jev(Config{MaxRequestBytes: 200})
	_, e := c.Encode(Request{Input: strings.Repeat("x", 200), Questions: []Question{{Name: "p", Type: Predicate}}})
	if e == nil {
		t.Fatal("budget bypass")
	}
	c, _ = Jev(Config{Model: strings.Repeat("m", 1000), MaxRequestBytes: 300})
	if _, e := c.Encode(Request{Input: "x", Questions: []Question{{Name: "p", Type: Predicate}}}); e == nil {
		t.Fatal("final wire budget bypass")
	}
	c, _ = Jev(Config{})
	r := Request{Input: json.RawMessage(`{"x":1}`), Questions: []Question{{Name: "p", Type: Predicate}}}
	if _, e := c.Encode(r); e != nil {
		t.Fatal(e)
	}
}
