package ftm

import "testing"

func TestNamespaceSignVerifyApply(t *testing.T) {
    ns := NewNamespace("dataset-key")
    signed := ns.Sign("p1")
    if signed == "p1" || signed == "" { t.Fatalf("expected signed id, got: %s", signed) }
    if !ns.Verify(signed) { t.Fatalf("verify failed: %s", signed) }

    // Apply to entity and referenced ids
    m, err := NewModel("../schema")
    if err != nil { t.Fatalf("NewModel: %v", err) }
    pass := m.Get("Passport")
    if pass == nil { t.Skip("Passport schema missing") }
    p := NewEntityProxy(pass, "doc1")
    // holder is entity reference
    _ = p.Add("holder", []string{"p1"}, true, "")
    sp := ns.Apply(p, false)
    if !ns.Verify(sp.ID) { t.Fatalf("verify applied entity id failed: %s", sp.ID) }
    vals := sp.Get("holder", false)
    if len(vals) == 0 || !ns.Verify(vals[0]) { t.Fatalf("verify applied holder failed: %v", vals) }
}
