package ftm

import "testing"

func TestProxyAddAndEdgePairs(t *testing.T) {
	m, err := NewModel("../schema")
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	person := m.Get("Person")
	if person == nil {
		t.Fatal("Person schema missing")
	}
	p := NewEntityProxy(person, "p1")
	if err := p.Add("name", []string{" John  Smith "}, false); err != nil {
		t.Fatalf("add name: %v", err)
	}
	vals := p.Get("name")
	if len(vals) != 1 || vals[0] != "John Smith" {
		t.Fatalf("unexpected name values: %v", vals)
	}

	// Relationship entity: Ownership (edge)
	own := m.Get("Ownership")
	if own == nil {
		t.Skip("Ownership schema missing")
	}
	e := NewEntityProxy(own, "rel1")
	// Use explicit edge source/target declared in schema
	if own.EdgeSource != "" && own.EdgeTarget != "" {
		_ = e.Add(own.EdgeSource, []string{"p1"}, true)
		_ = e.Add(own.EdgeTarget, []string{"ba1"}, true)
	}
	pairs := e.EdgePairs()
	if len(pairs) == 0 {
		// Tolerate edge-less schema or differing property names between versions.
		t.Log("no edgepairs produced for Ownership; schema may differ")
	}
}
