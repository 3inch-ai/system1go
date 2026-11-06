package system1go

import (
	"encoding/json"
	"math"
	"reflect"
)

func probability(v *float64) bool {
	return v != nil && !math.IsNaN(*v) && !math.IsInf(*v, 0) && *v >= 0 && *v <= 1
}
func (c *Client) decodeAnswer(raw json.RawMessage, q Question) Answer {
	out := Answer{Status: Invalid, Type: q.Type, Raw: raw, Issues: []string{"missing, malformed, or inconsistent answer"}}
	var a struct {
		Type          string
		Name          *string
		Noul          *float64
		Probability   *float64
		Choice        any
		Confidence    *float64
		Probabilities json.RawMessage
	}
	if json.Unmarshal(raw, &a) != nil {
		return out
	}
	if c.config.Protocol == OpenAIDecisions && (a.Name == nil || *a.Name != q.Name) {
		return out
	}
	if c.config.Protocol == OpenAIDecisions && a.Type == "refusal" {
		out.Status = Refused
		out.Issues = nil
		return out
	}
	if q.Type == Predicate && ((a.Type == "noul" && c.config.Protocol == SystemOne) || (a.Type == "predicate" && c.config.Protocol == OpenAIDecisions)) {
		p := a.Noul
		if c.config.Protocol == OpenAIDecisions {
			p = a.Probability
		}
		if probability(p) {
			out.Status = OK
			out.Probability = p
		}
		return out
	}
	if q.Type == Score {
		return c.decodeScore(raw, q, out)
	}
	if q.Type != Choice || a.Type != "choice" || !probability(a.Confidence) {
		return out
	}
	var entries []struct {
		Value       any
		Probability *float64
	}
	if c.config.Protocol == OpenAIDecisions {
		if json.Unmarshal(a.Probabilities, &entries) != nil {
			return out
		}
	} else {
		var m map[string]*float64
		if json.Unmarshal(a.Probabilities, &m) != nil {
			return out
		}
		for k, v := range m {
			entries = append(entries, struct {
				Value       any
				Probability *float64
			}{k, v})
		}
	}
	seen := map[any]bool{}
	sum := 0.0
	selected := false
	for _, o := range q.Choices {
		if reflect.DeepEqual(o.Value, a.Choice) {
			selected = true
		}
	}
	if !selected || len(entries) != len(q.Choices) {
		return out
	}
	for _, e := range entries {
		if !probability(e.Probability) {
			return out
		}
		allowed := false
		for _, o := range q.Choices {
			if reflect.DeepEqual(o.Value, e.Value) {
				allowed = true
			}
		}
		if !allowed || seen[e.Value] {
			return out
		}
		seen[e.Value] = true
		sum += *e.Probability
		out.Probabilities = append(out.Probabilities, Distribution{Value: e.Value, Probability: *e.Probability})
	}
	if math.Abs(sum-1) > 0.0100000001 {
		return out
	}
	selectedP := -1.0
	for _, e := range out.Probabilities {
		if reflect.DeepEqual(e.Value, a.Choice) {
			selectedP = e.Probability
		}
	}
	for _, e := range out.Probabilities {
		if e.Probability > selectedP+0.01 {
			return out
		}
	}
	out.Status = OK
	out.Confidence = a.Confidence
	out.Choice = a.Choice
	return out
}
