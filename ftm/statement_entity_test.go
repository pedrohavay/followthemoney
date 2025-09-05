package ftm

import "testing"

func TestStatementEntityBaseIDValue(t *testing.T) {
    m, err := NewModel("../schema")
    if err != nil {
        t.Fatalf("NewModel: %v", err)
    }
    sc := m.Get("Person")
    if sc == nil {
        t.Skip("Person schema not found")
    }
    se, err := NewStatementEntity(m, "dsX", "Person", "pX")
    if err != nil {
        t.Fatalf("NewStatementEntity: %v", err)
    }
    _ = se.Add(m, "name", "Alice", "", "", "t", "2025-01-02")
    out := se.Statements()
    found := false
    for _, s := range out {
        if s.Prop == BaseID {
            found = true
            if s.Value != "pX" {
                t.Fatalf("BaseID value should be entity id, got: %s", s.Value)
            }
            if s.PropType != BaseID {
                t.Fatalf("BaseID prop_type should be 'id', got: %s", s.PropType)
            }
        }
    }
    if !found {
        t.Fatalf("BaseID statement not found")
    }
}

