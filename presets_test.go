package system1go

import (
	"strings"
	"testing"
)

func TestPresetsAndCapabilities(t *testing.T) {
	for _, item := range []struct {
		f           func(Config) (*Client, error)
		path, model string
		min         int
	}{{Jev, "systemone", "jev-latest", 1}, {OpenAI, "decisions", "gpt-6-luna", 2}, {H2O, "systemone", "h2oai/h2o-lightning-4b", 1}} {
		c, e := item.f(Config{})
		if e != nil {
			t.Fatal(e)
		}
		if !strings.HasSuffix(c.endpoint, "/v1/"+item.path) || c.config.Model != item.model || c.Capabilities().MinChoices != item.min {
			t.Fatalf("%+v", c)
		}
	}
}
func TestStructuredLegend(t *testing.T) {
	c, _ := NewClient(Config{Protocol: SystemOne, Model: "m", BaseURL: "https://example.test"})
	r := Request{Input: "x", Questions: []Question{{Name: "s", Type: Score, Levels: []Level{{Label: "l", Description: map[string]any{"threshold": 2}}}}}}
	snap, e := c.snapshot(r)
	if e != nil {
		t.Fatal(e)
	}
	d, e := c.Decode([]byte(`{"answers":{"s":{"type":"score","score":0,"confidence":1,"legend":{"0":{"description":{"threshold":2},"label":"l"}},"probabilities":{"0":1}}}}`), snap.Questions)
	if e != nil || d.Answers["s"].Status != OK {
		t.Fatalf("%+v %v", d, e)
	}
}
