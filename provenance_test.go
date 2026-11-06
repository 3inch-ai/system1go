package system1go

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfidenceProvenance(t *testing.T) {
	c, _ := NewClient(Config{Protocol: SystemOne, Provider: "deployment", Model: "m", BaseURL: "https://example.test"})
	d, e := c.Decode([]byte(`{"answers":{"c":{"type":"choice","choice":"a","confidence":0.4,"probabilities":{"a":1}}}}`), []Question{{Name: "c", Type: Choice, Choices: []Option{{Value: "a"}}}})
	if e != nil {
		t.Fatal(e)
	}
	if d.Protocol != SystemOne || d.Provider != "deployment" || d.Answers["c"].ConfidenceSource != SystemOne {
		t.Fatalf("missing provenance %+v", d)
	}
}
func TestInvalidJSONDiagnostics(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "private malformed") }))
	defer s.Close()
	c, _ := Jev(Config{BaseURL: s.URL})
	_, err := c.Decide(context.Background(), Request{Input: "x", Questions: []Question{{Name: "p", Type: Predicate}}})
	e, ok := err.(*Error)
	if !ok || e.Code != "invalid_response" || string(e.Diagnostics()) != "private malformed" {
		t.Fatal(err)
	}
}
