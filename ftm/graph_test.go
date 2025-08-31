package ftm

import "testing"

func TestGraphEntityAndValueEdges(t *testing.T) {
	m, err := NewModel("../schema")
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	ps := m.Get("Person")
	if ps == nil {
		t.Skip("Person schema missing")
	}
	e := NewEntityProxy(ps, "p1")
	_ = e.Add("name", []string{"John Smith"}, false, "")

	g := NewGraph(nil)
	g.Add(e)
	if len(g.Nodes()) < 2 {
		t.Fatalf("expected at least 2 nodes (entity + value), got %d", len(g.Nodes()))
	}
	if len(g.Edges()) < 1 {
		t.Fatalf("expected at least 1 edge, got %d", len(g.Edges()))
	}
}
