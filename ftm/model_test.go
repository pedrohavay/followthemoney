package ftm

import (
    "os"
    "testing"
)

func TestDefaultLoadsLocalSchema(t *testing.T) {
    _ = os.Unsetenv("FTM_MODEL_PATH")
    // Use explicit relative path so tests are independent of CWD
    m, err := NewModel("../schema")
    if err != nil { t.Fatalf("NewModel: %v", err) }
    if m == nil { t.Fatalf("nil model") }
    // Basic sanity: common schemata should exist
    if m.Get("Person") == nil {
        t.Fatalf("expected Person schema to be present")
    }
    if m.Get("BankAccount") == nil {
        t.Fatalf("expected BankAccount schema to be present")
    }
    p := m.Get("Person").Get("name")
    if p == nil || p.Type.Name() != "name" {
        t.Fatalf("expected Person.name to be type 'name', got: %#v", p)
    }
}

func TestNewModelFromPath(t *testing.T) {
    m, err := NewModel("../schema")
    if err != nil { t.Fatalf("NewModel: %v", err) }
    if m.Get("Organization") == nil { t.Fatalf("expected Organization schema") }
}
