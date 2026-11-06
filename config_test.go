package system1go

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestConfigurationRejectsIgnoredControls(t *testing.T) {
	for _, cfg := range []Config{{MaxQuestions: 1}, {MaxChoices: 3}, {MaxLevels: 3}, {ExtraBodyKeys: []string{"ignored"}}} {
		cfg.Protocol = OpenAIDecisions
		cfg.Model = "m"
		cfg.BaseURL = "https://example.test"
		if _, e := NewClient(cfg); e == nil {
			t.Errorf("ignored config accepted %+v", cfg)
		}
	}
}
func TestMalformedModelMetadata(t *testing.T) {
	c, _ := Jev(Config{})
	d, e := c.Decode([]byte(`{"model":42,"answers":{"p":{"type":"noul","noul":0.5}}}`), []Question{{Name: "p", Type: Predicate}})
	if e != nil || len(d.Issues) == 0 || d.Answers["p"].Status != OK {
		t.Fatalf("%+v %v", d, e)
	}
}
func TestLargeStructuredLegend(t *testing.T) {
	c, _ := Jev(Config{})
	r := Request{Input: "x", Questions: []Question{{Name: "s", Type: Score, Levels: []Level{{Label: "l", Description: map[string]any{"n": uint64(9007199254740993)}}}}}}
	r, e := c.snapshot(r)
	if e != nil {
		t.Fatal(e)
	}
	b := []byte(fmt.Sprintf(`{"answers":{"s":{"type":"score","score":0,"confidence":1,"legend":{"0":{"label":"l","description":{"n":%s}}},"probabilities":{"0":1}}}}`, json.Number("9007199254740993")))
	d, e := c.Decode(b, r.Questions)
	if e != nil || d.Answers["s"].Status != OK {
		t.Fatalf("%+v %v", d, e)
	}
}
