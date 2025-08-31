package ftm

import (
	"testing"
)

func TestEmailCleaningIDNA(t *testing.T) {
	e := NewEmailType()
	cleaned, ok := e.Clean("John <j.smith@bücher.de>", false, "", nil)
	if !ok || cleaned != "j.smith@xn--bcher-kva.de" {
		t.Fatalf("email clean failed: %v %v", ok, cleaned)
	}
}

func TestCountryAndLanguageClean(t *testing.T) {
	c := NewCountryType()
	out, ok := c.Clean("DE", false, "", nil)
	if !ok || out != "de" {
		t.Fatalf("country clean failed: %v %v", ok, out)
	}

	l := NewLanguageType()
	out, ok = l.Clean("DEU", false, "", nil)
	if !ok || out != "deu" {
		t.Fatalf("language clean failed: %v %v", ok, out)
	}
}

func TestIdentifierIBANAndBIC(t *testing.T) {
	idt := NewIdentifierType()
	iban, ok := idt.Clean("DE44 5001 0517 5407 3249 31", false, "iban", nil)
	if !ok || iban != "DE44500105175407324931" {
		t.Fatalf("iban clean failed: %v %v", ok, iban)
	}
	bic, ok := idt.Clean("DEUTDEFF", false, "bic", nil)
	if !ok || bic != "DEUTDEFF" {
		t.Fatalf("bic clean failed: %v %v", ok, bic)
	}
}

func TestURLNormalizationAndCompare(t *testing.T) {
	u := NewURLType()
	a, ok := u.Clean("Example.com/Path?b=2&a=1#frag", false, "", nil)
	if !ok {
		t.Fatalf("url clean failed")
	}
	b, ok := u.Clean("http://example.com/Path?a=1&b=2", false, "", nil)
	if !ok {
		t.Fatalf("url clean failed (b)")
	}
	if u.Compare(a, b) < 1.0 {
		t.Fatalf("url compare expected equal: %s vs %s", a, b)
	}
}

func TestNameAndAddressCompare(t *testing.T) {
	n := NewNameType()
	if n.Compare("John Smith", "J. Smith") <= 0.0 {
		t.Fatalf("name compare too low")
	}

	a := NewAddressType()
	if a.Compare("Main St., Apt 2", "Main St Apt 2") <= 0.0 {
		t.Fatalf("address compare too low")
	}
}

func TestPhoneCleanWithCountryHint(t *testing.T) {
	m, err := NewModel("../schema")
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	ps := m.Get("Person")
	if ps == nil {
		t.Skip("Person schema missing")
	}
	p := NewEntityProxy(ps, "p1")
	// Add nationality first, so phone has country hint
	if err := p.Add("nationality", []string{"de"}, false, ""); err != nil {
		t.Fatalf("add nationality: %v", err)
	}
	// Now add phone
	ph := NewPhoneType()
	out, ok := ph.Clean("030 1234567", false, "", p)
	if !ok || out == "" {
		t.Fatalf("phone clean failed: %v %v", ok, out)
	}
	if out[0] != '+' {
		t.Fatalf("expected E.164, got: %s", out)
	}
}

func TestChecksumValidation(t *testing.T) {
	cs := NewChecksumType()
	if cs.Validate("DEADbeef") {
		t.Fatalf("short checksum should be invalid")
	}
	if !cs.Validate("0123456789abcdef0123456789abcdef01234567") {
		t.Fatalf("valid checksum failed")
	}
}

func TestJsonType(t *testing.T) {
	jt := NewJsonType()
	if !jt.Validate(`{"a":1}`) {
		t.Fatalf("json validate failed")
	}
	// raw string input should be JSON-string-encoded on clean
	out, ok := jt.Clean("hello", false, "", nil)
	if !ok || out != "\"hello\"" {
		t.Fatalf("json clean string failed: %v %v", ok, out)
	}
}
