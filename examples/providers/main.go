// Inspect provider-specific wire shapes offline. No credentials or HTTP required.
package main

import (
	"fmt"
	s "github.com/3inch-ai/system1go"
	"log"
)

func main() {
	r := s.Request{Input: map[string]any{"message": "Duplicate charge"}, Questions: []s.Question{{Name: "billing", Type: s.Predicate, Instructions: "Is this about billing?"}, {Name: "team", Type: s.Choice, Instructions: "Which team?", Choices: []s.Option{{Value: "billing"}, {Value: "shipping"}}}, {Name: "urgency", Type: s.Score, Instructions: "How urgent?", Levels: []s.Level{{Label: "low"}, {Label: "high"}}}}}
	for _, p := range []struct {
		name string
		new  func(s.Config) (*s.Client, error)
	}{{"Jev", s.Jev}, {"OpenAI", s.OpenAI}, {"H2O native", s.H2O}} {
		c, e := p.new(s.Config{})
		if e != nil {
			log.Fatal(e)
		}
		b, e := c.Encode(r)
		if e != nil {
			log.Fatal(e)
		}
		fmt.Printf("%s: %s\n", p.name, b)
	}
}
