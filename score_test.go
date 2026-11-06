package system1go

import (
	"encoding/json"
	"testing"
)

func TestScoreLabelPresence(t *testing.T) {
	c, err := NewClient(Config{Protocol: OpenAIDecisions, Model: "m", BaseURL: "https://example.test"})
	if err != nil {
		t.Fatal(err)
	}
	q := Question{Name: "s", Type: Score, Instructions: "rate", Levels: []Level{{Label: ""}, {Label: "high"}}}
	for _, tc := range []struct {
		name   string
		label  string
		status Status
	}{
		{name: "missing", label: "", status: Invalid},
		{name: "null", label: `,"label":null`, status: Invalid},
		{name: "explicitly empty", label: `,"label":""`, status: OK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := `{"answers":[{"name":"s","type":"score","score":0.8,"confidence":0.5,"probabilities":[{"value":0` + tc.label + `,"probability":0.2},{"value":1,"label":"high","probability":0.8}]}]}`
			d, err := c.Decode([]byte(raw), []Question{q})
			if err != nil {
				t.Fatal(err)
			}
			if got := d.Answers["s"].Status; got != tc.status {
				t.Fatalf("status = %s, want %s", got, tc.status)
			}
		})
	}
}

func TestScore(t *testing.T) {
	for _, p := range []Protocol{SystemOne, OpenAIDecisions} {
		t.Run(string(p), func(t *testing.T) {
			c, _ := NewClient(Config{Protocol: p, Model: "m", BaseURL: "https://example.test"})
			q := Question{Name: "s", Type: Score, Instructions: "rate", Levels: []Level{{Label: "low", Description: "not urgent"}, {Label: "high"}}}
			b, e := c.Encode(Request{Input: "x", Questions: []Question{q}})
			if e != nil {
				t.Fatal(e)
			}
			var body map[string]any
			json.Unmarshal(b, &body)
			if p == SystemOne {
				v := body["questions"].(map[string]any)["s"].(map[string]any)
				if v["criteria"] == nil {
					t.Fatalf("missing criteria: %s", b)
				}
			} else {
				v := body["questions"].([]any)[0].(map[string]any)
				if v["levels"] == nil {
					t.Fatalf("missing levels: %s", b)
				}
			}
			raw := `{"answers":{"s":{"type":"score","score":0.7,"confidence":0.4,"legend":{"0":{"label":"low","description":"not urgent"},"1":"high"},"probabilities":{"0":0.3,"1":0.7}}}}`
			if p == OpenAIDecisions {
				raw = `{"answers":[{"name":"s","type":"score","score":0.7,"confidence":0.4,"probabilities":[{"value":0,"label":"low","probability":0.3},{"value":1,"label":"high","probability":0.7}]}]}`
			}
			d, e := c.Decode([]byte(raw), []Question{q})
			if e != nil {
				t.Fatal(e)
			}
			a := d.Answers["s"]
			if a.Status != OK || a.Score == nil || *a.Score != 0.7 || a.ComputedScore == nil || len(a.Levels) != 2 {
				t.Fatalf("%+v", a)
			}
		})
	}
}
