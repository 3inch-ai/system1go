package system1go

import (
	"bytes"
	"encoding/json"
	"reflect"
	"unicode/utf8"
)

// snapshot invokes custom MarshalJSON exactly once, then validates the materialized JSON.
func (c *Client) snapshot(r Request) (Request, error) {
	if !validStrings(reflect.ValueOf(r), 0) {
		return Request{}, fail("invalid_request")
	}
	b, e := json.Marshal(r)
	if e != nil {
		return Request{}, fail("invalid_request")
	}
	if int64(len(b)) > c.config.MaxRequestBytes {
		return Request{}, fail("request_too_large")
	}
	var v Request
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if decoder.Decode(&v) != nil {
		return Request{}, fail("invalid_request")
	}
	return v, nil
}
func content(v any, nullable bool) bool {
	switch v.(type) {
	case string, map[string]any, []any:
		return true
	case nil:
		return nullable
	}
	return false
}
func text(v any, max int) bool {
	s, ok := v.(string)
	return ok && utf8.ValidString(s) && utf8.RuneCountInString(s) <= max
}
func (c *Client) validate(r Request) error {
	bad := func(q, f string) error {
		return &Error{Code: "unsupported", Question: q, Feature: f, Protocol: c.config.Protocol, Provider: c.config.Provider}
	}
	open := c.config.Protocol == OpenAIDecisions
	min := 1
	if open {
		min = 2
	}
	maxQ := c.config.MaxQuestions
	if open {
		maxQ = 400
	}
	if len(r.Questions) == 0 {
		return fail("invalid_request")
	}
	if maxQ > 0 && len(r.Questions) > maxQ {
		return bad("", "questions")
	}
	if !content(r.Input, false) {
		return fail("invalid_request")
	}
	if open {
		if a, ok := r.Input.([]any); ok {
			for _, v := range a {
				if _, ok := v.(map[string]any); !ok {
					return bad("", "input_json")
				}
			}
		}
		if _, ok := r.Input.(string); !ok {
			b, _ := json.Marshal(r.Input)
			if len(b) > 262144 {
				return bad("", "input_json")
			}
			count := 0
			if !jsonLimits(r.Input, 1, 32, &count, 10000) {
				return bad("", "input_json")
			}
		}
	}
	count := 0
	if !jsonLimits(r.Input, 1, 64, &count, 0) {
		return bad("", "json_depth")
	}
	names := map[string]bool{}
	for _, q := range r.Questions {
		if names[q.Name] {
			return fail("invalid_request")
		}
		names[q.Name] = true
		if !content(q.Instructions, true) {
			return bad(q.Name, "instructions")
		}
		if open && (!text(q.Instructions, 1048576) || !text(q.Name, 1048576)) {
			return bad(q.Name, "instructions")
		}
		switch q.Type {
		case Predicate:
			if len(q.Choices) > 0 || len(q.Levels) > 0 {
				return bad(q.Name, "fields")
			}
			if q.Criteria != nil {
				if open {
					return bad(q.Name, "criteria")
				}
				for k, v := range q.Criteria {
					if (k != "true" && k != "false") || !content(v, true) {
						return bad(q.Name, "criteria")
					}
				}
			}
		case Choice:
			if q.Criteria != nil || len(q.Levels) > 0 {
				return bad(q.Name, "fields")
			}
			max := c.config.MaxChoices
			if max == 0 || open {
				max = 255
			}
			if len(q.Choices) < min || len(q.Choices) > max {
				return bad(q.Name, "choices")
			}
			seen := map[any]bool{}
			for _, o := range q.Choices {
				switch v := o.Value.(type) {
				case string:
					if open && !text(v, 1048576) {
						return bad(q.Name, "choice.value")
					}
				case bool:
					if !open {
						return bad(q.Name, "booleanChoices")
					}
				default:
					return bad(q.Name, "choice.value")
				}
				if seen[o.Value] {
					return fail("invalid_request")
				}
				seen[o.Value] = true
				if !content(o.Description, true) || open && o.Description != nil && !text(o.Description, 1048576) {
					return bad(q.Name, "description")
				}
			}
		case Score:
			if q.Criteria != nil || len(q.Choices) > 0 {
				return bad(q.Name, "fields")
			}
			max := c.config.MaxLevels
			if max == 0 || open {
				max = 10
			}
			if len(q.Levels) < min || len(q.Levels) > max {
				return bad(q.Name, "levels")
			}
			for _, l := range q.Levels {
				if !content(l.Description, true) || open && (!text(l.Label, 1048576) || l.Description != nil && !text(l.Description, 1048576)) {
					return bad(q.Name, "levels")
				}
			}
		default:
			return bad(q.Name, "type")
		}
	}
	for k, v := range r.Options {
		if !open {
			if k == "model" || k == "state" || k == "questions" || k == "assets" {
				return bad("", k)
			}
			allowed := false
			for _, key := range c.config.ExtraBodyKeys {
				if key == k {
					allowed = true
				}
			}
			if !allowed {
				return bad("", k)
			}
			continue
		}
		switch k {
		case "prompt_cache_key":
			if v != nil && !text(v, 64) {
				return bad("", k)
			}
		case "safety_identifier":
			if v != nil && !text(v, 128) {
				return bad("", k)
			}
		case "inputTextCacheBreakpoint":
			b, ok := v.(bool)
			if !ok {
				return bad("", k)
			}
			if b {
				if _, ok := r.Input.(string); !ok {
					return bad("", k)
				}
			}
		case "prompt_cache_options":
			if v == nil {
				continue
			}
			m, ok := v.(map[string]any)
			if !ok {
				return bad("", k)
			}
			for key, v := range m {
				if key == "mode" && (v == nil || v == "explicit" || v == "implicit") {
					continue
				}
				if key == "ttl" && v == "30m" {
					continue
				}
				return bad("", k)
			}
		default:
			return bad("", k)
		}
	}
	return nil
}
func jsonLimits(v any, depth, maxDepth int, count *int, maxCount int) bool {
	*count++
	if maxCount > 0 && *count > maxCount {
		return false
	}
	switch v := v.(type) {
	case map[string]any:
		if depth > maxDepth {
			return false
		}
		for _, x := range v {
			if !jsonLimits(x, depth+1, maxDepth, count, maxCount) {
				return false
			}
		}
	case []any:
		if depth > maxDepth {
			return false
		}
		for _, x := range v {
			if !jsonLimits(x, depth+1, maxDepth, count, maxCount) {
				return false
			}
		}
	}
	return true
}
