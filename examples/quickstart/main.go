// This runnable example uses synthetic loopback HTTP, never a real model.
package main

import (
	"context"
	"fmt"
	s "github.com/3inch-ai/system1go"
	"log"
	"net/http"
	"net/http/httptest"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"model":"synthetic","answers":{"billing":{"type":"noul","noul":0.95}}}`)
	}))
	defer server.Close()
	client, err := s.Jev(s.Config{BaseURL: server.URL})
	if err != nil {
		log.Fatal(err)
	}
	d, err := client.Decide(context.Background(), s.Request{Input: "Duplicate charge", Questions: []s.Question{{Name: "billing", Type: s.Predicate, Instructions: "Is this about billing?"}}})
	if err != nil {
		log.Fatal(err)
	}
	a := d.Answers["billing"]
	if a.Status != s.OK {
		log.Fatal("invalid synthetic answer")
	}
	fmt.Printf("%s: %.2f (synthetic HTTP, no inference)\n", a.Status, *a.Probability)
}
