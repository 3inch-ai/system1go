package system1go

import (
	"testing"
)

func TestInvalidUTF8IsNotReplaced(t *testing.T) {
	c, _ := Jev(Config{})
	for _, input := range []any{string([]byte{0xff}), map[string]any{string([]byte{0xff}): "x"}, map[string]any{"nested": []any{string([]byte{0xfe})}}} {
		if b, e := c.Encode(Request{Input: input, Questions: []Question{{Name: "p", Type: Predicate}}}); e == nil {
			t.Errorf("invalid UTF-8 silently changed: %s", b)
		}
	}
	q := Question{Name: string([]byte{0xff}), Type: Predicate}
	if _, e := c.Encode(Request{Input: "x", Questions: []Question{q}}); e == nil {
		t.Error("invalid name replaced")
	}
	deep := any("x")
	for i := 0; i < 65; i++ {
		deep = []any{deep}
	}
	if _, e := c.Encode(Request{Input: deep, Questions: []Question{{Name: "p", Type: Predicate}}}); e == nil {
		t.Error("deep native state accepted")
	}
}
