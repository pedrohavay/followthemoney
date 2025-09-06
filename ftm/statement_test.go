package ftm

import (
	"bytes"
	"sort"
	"strings"
	"testing"
)

func TestStatementsFromEntityAndAggregate(t *testing.T) {
	m, err := NewModel("../schema")
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	sc := m.Get("Person")
	if sc == nil {
		t.Skip("Person schema not found")
	}
	e := NewEntityProxy(sc, "p1")
	_ = e.Add("name", []string{"John Smith"}, false)
	_ = e.Add("nationality", []string{"de"}, false)

	st := StatementsFromEntity(e, "ds1", "2024-01-01", "", false, "test")
	if len(st) < 3 {
		t.Fatalf("expected >= 3 statements (base + props), got %d", len(st))
	}
	// ensure ids are set
	for _, s := range st {
		if s.ID == "" {
			t.Fatalf("statement without id: %#v", s)
		}
		if s.PropType == "" {
			t.Fatalf("statement without prop_type: %#v", s)
		}
	}

	// JSONL round-trip
	buf := bytes.Buffer{}
	if err := WriteStatementsJSONL(&buf, st); err != nil {
		t.Fatalf("write jsonl: %v", err)
	}
	var back []Statement
	if err := ReadStatementsJSONL(&buf, func(s Statement) error { back = append(back, s); return nil }); err != nil {
		t.Fatalf("read jsonl: %v", err)
	}
	if len(back) != len(st) {
		t.Fatalf("jsonl round-trip length mismatch: %d vs %d", len(back), len(st))
	}

	// Aggregation on sorted statements
	sort.Slice(back, func(i, j int) bool { return back[i].GroupKey() < back[j].GroupKey() })
	es := AggregateSortedStatements(m, back)
	if len(es) != 1 {
		t.Fatalf("expected 1 aggregated entity, got %d", len(es))
	}
	ag := es[0]
	if ag.ID != "p1" {
		t.Fatalf("aggregated id mismatch: %s", ag.ID)
	}
	if ag.First("name") != "John Smith" {
		t.Fatalf("name lost in aggregate")
	}
}

func TestAggregateCanonicalIDCollapse(t *testing.T) {
	m, err := NewModel("../schema")
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	// Two statements for different entity IDs but same canonical
	s1 := Statement{EntityID: "a", CanonicalID: "X", Prop: BaseID, Schema: "Person", Value: "a", Dataset: "ds"}
	s1.MakeKey()
	s2 := Statement{EntityID: "b", CanonicalID: "X", Prop: "name", Schema: "Person", Value: "John", Dataset: "ds"}
	s2.MakeKey()
	st := []Statement{s1, s2}
	es := AggregateSortedStatements(m, st)
	if len(es) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(es))
	}
	if es[0].ID != "X" {
		t.Fatalf("expected canonical id 'X', got %s", es[0].ID)
	}
}

func TestStatementsCSVAndMsgpackRoundTrip(t *testing.T) {
	m, err := NewModel("../schema")
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	sc := m.Get("Person")
	if sc == nil {
		t.Skip("Person schema not found")
	}
	e := NewEntityProxy(sc, "p2")
	_ = e.Add("name", []string{"Maria"}, false)
	_ = e.Add("nationality", []string{"br"}, false)
	st := StatementsFromEntity(e, "ds2", "2025-01-01", "", false, "test")

	// CSV write/read
	csvbuf := bytes.Buffer{}
	if err := WriteStatementsCSV(&csvbuf, st); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	var backCSV []Statement
	if err := ReadStatementsCSV(strings.NewReader(csvbuf.String()), func(s Statement) error { backCSV = append(backCSV, s); return nil }); err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if len(backCSV) != len(st) {
		t.Fatalf("csv round-trip length mismatch: %d vs %d", len(backCSV), len(st))
	}
	// prop_type must be present after round-trip
	for _, s := range backCSV {
		if s.PropType == "" {
			t.Fatalf("csv round-trip missing prop_type for statement: %#v", s)
		}
	}

	// Msgpack write/read
	mpbuf := bytes.Buffer{}
	if err := WriteStatementsMsgpack(&mpbuf, st); err != nil {
		t.Fatalf("write msgpack: %v", err)
	}
	var backMP []Statement
	if err := ReadStatementsMsgpack(&mpbuf, func(s Statement) error { backMP = append(backMP, s); return nil }); err != nil {
		t.Fatalf("read msgpack: %v", err)
	}
	if len(backMP) != len(st) {
		t.Fatalf("msgpack round-trip length mismatch: %d vs %d", len(backMP), len(st))
	}
	for _, s := range backMP {
		if s.PropType == "" {
			t.Fatalf("msgpack round-trip missing prop_type for statement: %#v", s)
		}
	}
}

// BaseID semantics are tested in statement_entity_test.go

func TestStatementAggregatorStream(t *testing.T) {
	m, err := NewModel("../schema")
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	st := []Statement{
		{EntityID: "a", CanonicalID: "a", Prop: BaseID, Schema: "Person", Value: "a", Dataset: "ds"},
		{EntityID: "a", CanonicalID: "a", Prop: "name", Schema: "Person", Value: "Ana", Dataset: "ds"},
		{EntityID: "b", CanonicalID: "b", Prop: BaseID, Schema: "Person", Value: "b", Dataset: "ds"},
		{EntityID: "b", CanonicalID: "b", Prop: "name", Schema: "Person", Value: "Bob", Dataset: "ds"},
	}
	for i := range st {
		st[i].MakeKey()
	}
	agg := NewStatementAggregator(m)
	var out []*EntityProxy
	for i := range st {
		if ent := agg.Add(st[i]); ent != nil {
			out = append(out, ent)
		}
	}
	if ent := agg.Flush(); ent != nil {
		out = append(out, ent)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(out))
	}
}
