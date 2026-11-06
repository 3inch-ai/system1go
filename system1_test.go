package system1go

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNativePredicate(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b map[string]any
		json.NewDecoder(r.Body).Decode(&b)
		if r.URL.Path != "/prefix/v1/systemone" || b["state"] != "hello" {
			t.Errorf("wire: %v %s", b, r.URL)
		}
		q := b["questions"].(map[string]any)["p"].(map[string]any)
		if q["type"] != "noul" {
			t.Error(q)
		}
		w.Write([]byte(`{"model":"actual","answers":{"p":{"type":"noul","noul":0.8}},"attention_mode":"native"}`))
	}))
	defer s.Close()
	c, e := NewClient(Config{Protocol: SystemOne, BaseURL: s.URL + "/prefix/v1/", Model: "requested"})
	if e != nil {
		t.Fatal(e)
	}
	d, e := c.Decide(context.Background(), Request{Input: "hello", Questions: []Question{{Name: "p", Type: Predicate, Instructions: "true?"}}})
	if e != nil {
		t.Fatal(e)
	}
	if d.Answers["p"].Status != OK || *d.Answers["p"].Probability != 0.8 || d.ActualModel != "actual" || d.Usage != nil || len(d.Metadata) == 0 || len(d.Raw) == 0 {
		t.Fatalf("%+v", d)
	}
}
