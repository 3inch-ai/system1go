package system1go

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
)

func levelContent(l Level) any {
	if l.Description == nil {
		return l.Label
	}
	return map[string]any{"label": l.Label, "description": l.Description}
}
func (c *Client) decodeScore(raw json.RawMessage, q Question, out Answer) Answer {
	var a struct {
		Type          string
		Score         *float64
		Confidence    *float64
		Probabilities json.RawMessage
		Legend        map[string]any
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&a) != nil || a.Type != "score" || !probability(a.Confidence) || a.Score == nil || *a.Score < 0 || *a.Score > float64(len(q.Levels)-1) {
		return out
	}
	type entry struct {
		Value       int
		Label       string
		Probability *float64
	}
	entries := []entry{}
	if c.config.Protocol == SystemOne {
		var m map[string]*float64
		if json.Unmarshal(a.Probabilities, &m) != nil || len(a.Legend) != len(q.Levels) {
			return out
		}
		for i, l := range q.Levels {
			if !jsonEqual(a.Legend[strconv.Itoa(i)], levelContent(l)) {
				return out
			}
		}
		for k, p := range m {
			i, e := strconv.Atoi(k)
			if e != nil || strconv.Itoa(i) != k {
				return out
			}
			entries = append(entries, entry{Value: i, Probability: p})
		}
	} else {
		var wire []struct {
			Value       *int
			Label       *string
			Probability *float64
		}
		if json.Unmarshal(a.Probabilities, &wire) != nil {
			return out
		}
		for _, e := range wire {
			if e.Value == nil || e.Label == nil {
				return out
			}
			entries = append(entries, entry{Value: *e.Value, Label: *e.Label, Probability: e.Probability})
		}
	}
	if len(entries) != len(q.Levels) {
		return out
	}
	seen := map[int]bool{}
	sum, weighted := 0.0, 0.0
	for _, e := range entries {
		if e.Value < 0 || e.Value >= len(q.Levels) || seen[e.Value] || !probability(e.Probability) {
			return out
		}
		seen[e.Value] = true
		if c.config.Protocol == OpenAIDecisions && e.Label != q.Levels[e.Value].Label {
			return out
		}
		sum += *e.Probability
		weighted += float64(e.Value) * *e.Probability
		out.Probabilities = append(out.Probabilities, Distribution{Value: e.Value, Label: q.Levels[e.Value].Label, Probability: *e.Probability})
	}
	if math.Abs(sum-1) > 0.0100000001 || math.Abs(weighted-*a.Score) > 0.0200000001 {
		return out
	}
	out.Status = OK
	out.Score = a.Score
	out.ComputedScore = &weighted
	out.Confidence = a.Confidence
	out.Levels = q.Levels
	return out
}

func jsonEqual(a, b any) bool {
	x, e := json.Marshal(a)
	y, d := json.Marshal(b)
	return e == nil && d == nil && bytes.Equal(x, y)
}
