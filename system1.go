// Package system1go provides explicit, lossless System One and Decisions clients.
package system1go

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Protocol identifies a wire contract independently of provider or model.
type Protocol string

const (
	SystemOne       Protocol = "system-one"
	OpenAIDecisions Protocol = "openai-decisions"
)

// Kind is a common decision primitive.
type Kind string

const (
	Predicate Kind = "predicate"
	Choice    Kind = "choice"
	Score     Kind = "score"
)

// Status distinguishes usable answers, explicit refusals and invalid answers.
type Status string

const (
	OK      Status = "ok"
	Refused Status = "refused"
	Invalid Status = "invalid"
)

// Question is a named primitive. Only fields appropriate to Type may be set.
// Native Instructions may be nil or structured; OpenAI requires text.
type Question struct {
	Name         string
	Type         Kind
	Instructions any
	Criteria     map[string]any
	Choices      []Option
	Levels       []Level
}

// Option has a string or boolean Value. Native System One permits strings only.
type Option struct {
	Value       any `json:"value"`
	Description any `json:"description,omitempty"`
}

// Level is an ordered score criterion. Description may be structured natively.
type Level struct {
	Label       string `json:"label"`
	Description any    `json:"description,omitempty"`
}

// Request supplies shared text/JSON evidence and uniquely named ordered questions.
// Options accepts only explicitly implemented or declared provider fields.
type Request struct {
	Input     any
	Questions []Question
	Options   map[string]any
}

// Answer is usable only when Status is OK. Numeric pointers preserve absence.
// Confidence is provider-specific certainty, not a universal correctness probability.
// Score is original; ComputedScore is separate verification, never an overwrite.
type Answer struct {
	ConfidenceSource Protocol
	Status           Status
	Type             Kind
	Probability      *float64
	Choice           any
	Score            *float64
	ComputedScore    *float64
	Confidence       *float64
	Probabilities    []Distribution
	Levels           []Level
	Issues           []string
	Raw              json.RawMessage
}

// Distribution preserves typed choice values or zero-based integer score indices.
type Distribution struct {
	Value       any
	Label       string
	Probability float64
}

// Decision retains normalized answers and original response evidence.
// Usage is nil when absent; Raw and Metadata may contain sensitive input.
type Decision struct {
	Protocol       Protocol
	Provider       string
	Answers        map[string]Answer
	Raw            json.RawMessage
	Metadata       map[string]json.RawMessage
	Usage          json.RawMessage
	ActualModel    string
	RequestedModel string
	Issues         []string
	HTTP           HTTPMetadata
}

// HTTPMetadata includes only status and an allowlisted subset of response headers.
type HTTPMetadata struct {
	Status    int
	RequestID string
	Headers   http.Header
}

// Error is a safe ordinary error; explicit Diagnostics and Cause may be sensitive.
// Use errors.As to inspect Code, Question, Feature and HTTP status.
type Error struct {
	Protocol Protocol
	Provider string
	Code     string
	Question string
	Feature  string
	Status   int
	HTTP     HTTPMetadata
	raw      []byte
	cause    error
}

func (e *Error) Error() string { return "system1go: " + e.Code }

// Diagnostics explicitly returns potentially sensitive upstream bytes. Never log automatically.
func (e *Error) Diagnostics() []byte { return bytes.Clone(e.raw) }
func fail(code string) *Error        { return &Error{Code: code} }

// Config is snapshotted by NewClient. Zero timeout and byte budgets use safe defaults.
// MaxQuestions/MaxChoices/MaxLevels and ExtraBodyKeys apply to native deployments only.
// HTTPClient is copied; its shared Transport/Jar must obey normal net/http concurrency rules.
type Config struct {
	Protocol         Protocol
	Provider         string
	BaseURL          string
	Model            string
	APIKey           string
	Headers          http.Header
	HTTPClient       *http.Client
	Timeout          time.Duration
	MaxResponseBytes int64
	MaxRequestBytes  int64
	ExtraBodyKeys    []string
	MaxQuestions     int
	MaxChoices       int
	MaxLevels        int
}

// Client is safe for concurrent calls with independently owned request data.
type Client struct {
	config   Config
	endpoint string
	http     *http.Client
}

// NewClient validates configuration without network I/O or environment access.
func NewClient(c Config) (*Client, error) {
	if c.Protocol == OpenAIDecisions && (c.MaxQuestions != 0 || c.MaxChoices != 0 || c.MaxLevels != 0 || len(c.ExtraBodyKeys) > 0) {
		return nil, fail("invalid_request")
	}
	if c.Protocol != SystemOne && c.Protocol != OpenAIDecisions || c.Model == "" {
		return nil, fail("invalid_request")
	}
	u, e := url.Parse(c.BaseURL)
	if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fail("invalid_request")
	}
	p := strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(p, "/v1") {
		p += "/v1"
	}
	method := "systemone"
	if c.Protocol == OpenAIDecisions {
		method = "decisions"
	}
	u.Path = p + "/" + method
	u.RawPath = ""
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	if c.MaxResponseBytes == 0 {
		c.MaxResponseBytes = 8 << 20
	}
	if c.MaxRequestBytes == 0 {
		c.MaxRequestBytes = 32 << 20
	}
	if c.Timeout < 0 || c.MaxResponseBytes < 1 || c.MaxRequestBytes < 1 {
		return nil, fail("invalid_request")
	}
	headers := make(http.Header)
	for k, values := range c.Headers {
		if !validHeaderName(k) {
			return nil, fail("invalid_request")
		}
		for _, v := range values {
			if !validHeaderValue(v) {
				return nil, fail("invalid_request")
			}
			headers.Add(k, v)
		}
	}
	c.Headers = headers
	if len(c.Headers.Values("Authorization")) > 1 {
		return nil, fail("invalid_request")
	}
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	c.Headers.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		if len(c.Headers.Values("Authorization")) > 0 {
			return nil, fail("invalid_request")
		}
		c.Headers.Set("Authorization", "Bearer "+c.APIKey)
	}
	h := http.Client{}
	for name, values := range c.Headers {
		if !validHeaderName(name) {
			return nil, fail("invalid_request")
		}
		for _, v := range values {
			if !validHeaderValue(v) {
				return nil, fail("invalid_request")
			}
		}
	}
	if c.MaxResponseBytes > 1<<30 || c.MaxRequestBytes > 1<<30 || c.MaxQuestions < 0 || c.MaxChoices < 0 || c.MaxLevels < 0 {
		return nil, fail("invalid_request")
	}
	if c.Headers.Get("Host") != "" {
		return nil, fail("invalid_request")
	}
	if c.HTTPClient != nil {
		h = *c.HTTPClient
	}
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	c.ExtraBodyKeys = append([]string(nil), c.ExtraBodyKeys...)
	return &Client{c, u.String(), &h}, nil
}

// Encode performs pure protocol conversion without network access.
func (c *Client) Encode(r Request) ([]byte, error) {
	snap, e := c.snapshot(r)
	if e != nil {
		return nil, e
	}
	return c.encodeSnapshot(snap)
}
func (c *Client) encodeSnapshot(r Request) (encoded []byte, err error) {
	defer func() {
		if err == nil && int64(len(encoded)) > c.config.MaxRequestBytes {
			encoded = nil
			err = fail("request_too_large")
		}
	}()
	if e := c.validate(r); e != nil {
		return nil, e
	}
	qs := map[string]any{}
	ordered := []any{}
	for _, q := range r.Questions {
		t := q.Type
		if t == Predicate && c.config.Protocol == SystemOne {
			t = "noul"
		}
		b := map[string]any{"type": t, "instructions": q.Instructions}
		if q.Type == Score {
			if c.config.Protocol == SystemOne {
				criteria := []any{}
				for _, l := range q.Levels {
					criteria = append(criteria, levelContent(l))
				}
				b["criteria"] = criteria
			} else {
				b["levels"] = q.Levels
			}
		}
		if q.Type == Choice {
			if c.config.Protocol == SystemOne {
				criteria := map[string]any{}
				for _, v := range q.Choices {
					s, ok := v.Value.(string)
					if !ok {
						return nil, fail("unsupported")
					}
					criteria[s] = v.Description
				}
				b["criteria"] = criteria
			} else {
				b["choices"] = q.Choices
			}
		}
		if q.Type == Predicate && q.Criteria != nil {
			b["criteria"] = q.Criteria
		}
		qs[q.Name] = b
		ob := map[string]any{}
		for k, v := range b {
			ob[k] = v
		}
		ob["name"] = q.Name
		ordered = append(ordered, ob)
	}
	if c.config.Protocol == OpenAIDecisions {
		input := r.Input
		if flag, _ := r.Options["inputTextCacheBreakpoint"].(bool); flag {
			input = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_text", "text": input, "prompt_cache_breakpoint": map[string]any{"mode": "explicit"}}}}}
		} else if _, ok := input.(string); !ok {
			input = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_json", "data": input}}}}
		}
		body := map[string]any{"model": c.config.Model, "input": input, "questions": ordered}
		for k, v := range r.Options {
			if k != "inputTextCacheBreakpoint" {
				body[k] = v
			}
		}
		return json.Marshal(body)
	}
	body := map[string]any{"model": c.config.Model, "state": r.Input, "questions": qs}
	for k, v := range r.Options {
		body[k] = v
	}
	return json.Marshal(body)
}

// Decode preserves the original JSON and absent usage.
func (c *Client) Decode(raw []byte, questions []Question) (*Decision, error) {
	if e := c.validate(Request{Input: "", Questions: questions}); e != nil {
		return nil, e
	}
	var env map[string]json.RawMessage
	if json.Unmarshal(raw, &env) != nil || env == nil {
		return nil, fail("invalid_response")
	}
	var answers map[string]json.RawMessage
	if c.config.Protocol == OpenAIDecisions {
		var arr []json.RawMessage
		if json.Unmarshal(env["answers"], &arr) != nil || arr == nil {
			return nil, fail("invalid_response")
		}
		answers = map[string]json.RawMessage{}
		for i, q := range questions {
			if i < len(arr) {
				answers[q.Name] = arr[i]
			}
		}
	} else if json.Unmarshal(env["answers"], &answers) != nil || answers == nil {
		return nil, fail("invalid_response")
	}
	d := &Decision{Protocol: c.config.Protocol, Provider: c.config.Provider, Answers: map[string]Answer{}, Raw: bytes.Clone(raw), Metadata: map[string]json.RawMessage{}, Usage: env["usage"], RequestedModel: c.config.Model}
	if v, ok := env["model"]; ok {
		var model *string
		if json.Unmarshal(v, &model) != nil || model == nil {
			d.Issues = append(d.Issues, "response model is not a string")
		} else {
			d.ActualModel = *model
		}
	}
	for k, v := range env {
		if k != "answers" && k != "model" && k != "usage" {
			d.Metadata[k] = v
		}
	}
	for _, q := range questions {
		out := c.decodeAnswer(answers[q.Name], q)
		if c.config.Protocol == OpenAIDecisions {
			var items []struct{ Name *string }
			json.Unmarshal(env["answers"], &items)
			n := 0
			for _, item := range items {
				if item.Name != nil && *item.Name == q.Name {
					n++
				}
			}
			if n > 1 {
				out.Status = Invalid
				out.Issues = []string{"duplicate answer name"}
			}
		}
		if out.Status != Invalid {
			if out.Confidence != nil {
				out.ConfidenceSource = c.config.Protocol
			}
			out.Issues = nil
		}
		d.Answers[q.Name] = out
	}
	if len(answers) != len(questions) {
		d.Issues = append(d.Issues, "answer count differs")
	}
	if c.config.Protocol == SystemOne {
		for name := range answers {
			found := false
			for _, q := range questions {
				if name == q.Name {
					found = true
				}
			}
			if !found {
				d.Issues = append(d.Issues, "unexpected answer ID")
			}
		}
	} else {
		var arr []json.RawMessage
		json.Unmarshal(env["answers"], &arr)
		if len(arr) != len(questions) {
			d.Issues = append(d.Issues, "answer count differs")
		}
	}
	return d, nil
}

// Decide validates and snapshots the request, then makes exactly one HTTP request.
// The context and configured timeout cover response body consumption.
func (c *Client) Decide(ctx context.Context, r Request) (*Decision, error) {
	snap, e := c.snapshot(r)
	if e != nil {
		return nil, e
	}
	body, e := c.encodeSnapshot(snap)
	if e != nil {
		return nil, e
	}
	if int64(len(body)) > c.config.MaxRequestBytes {
		return nil, fail("request_too_large")
	}
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if e != nil {
		return nil, fail("invalid_request")
	}
	req.Header = c.config.Headers.Clone()
	resp, e := c.http.Do(req)
	if e != nil {
		return nil, transportError(ctx, e)
	}
	defer resp.Body.Close()
	meta := safeHTTP(resp)
	raw, e := io.ReadAll(io.LimitReader(resp.Body, c.config.MaxResponseBytes+1))
	if e != nil {
		return nil, transportError(ctx, e)
	}
	if int64(len(raw)) > c.config.MaxResponseBytes {
		return nil, &Error{Code: "response_too_large", HTTP: meta, Status: resp.StatusCode}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &Error{Code: statusCode(resp.StatusCode), Status: resp.StatusCode, HTTP: meta, raw: raw}
	}
	d, e := c.Decode(raw, snap.Questions)
	if e != nil {
		if se, ok := e.(*Error); ok {
			se.raw = raw
			se.HTTP = meta
			se.Status = resp.StatusCode
		}
		return nil, e
	}
	d.HTTP = meta
	return d, nil
}
