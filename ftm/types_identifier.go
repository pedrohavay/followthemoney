package ftm

import (
	"math/big"
	"regexp"
	"strings"
)

// IdentifierType with optional format validation (IBAN, LEI, etc.).
type IdentifierType struct{ BaseType }

func NewIdentifierType() *IdentifierType {
	return &IdentifierType{BaseType{name: "identifier", group: "identifiers", label: "Identifier", matchable: true, pivot: true, maxLength: 64}}
}
func (t *IdentifierType) Validate(value string) bool {
	_, ok := t.Clean(value, false, "", nil)
	return ok
}
func (t *IdentifierType) Clean(text string, _ bool, format string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	switch strings.ToLower(format) {
	case "iban":
		if iban := normalizeIBAN(s); iban != "" {
			return iban, true
		}
		return "", false
	case "lei":
		u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
		if regexp.MustCompile(`^[A-Z0-9]{20}$`).MatchString(u) {
			return u, true
		}
		return "", false
	case "bic":
		u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
		if regexp.MustCompile(`^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$`).MatchString(u) {
			return u, true
		}
		return "", false
	case "isin":
		u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
		if regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`).MatchString(u) {
			return u, true
		}
		return "", false
	case "figi":
		u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
		if regexp.MustCompile(`^[A-Z0-9]{12}$`).MatchString(u) {
			return u, true
		}
		return "", false
	case "ssn":
		digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
		if len(digits) == 9 {
			return digits, true
		}
		return "", false
	case "uscc":
		u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
		if regexp.MustCompile(`^[0-9A-Z]{18}$`).MatchString(u) {
			return u, true
		}
		return "", false
	case "inn":
		digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
		if len(digits) == 10 || len(digits) == 12 {
			return digits, true
		}
		return "", false
	case "ogrn":
		digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
		if len(digits) == 13 || len(digits) == 15 {
			return digits, true
		}
		return "", false
	case "uei":
		u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
		if regexp.MustCompile(`^[A-Z0-9]{12}$`).MatchString(u) {
			return u, true
		}
		return "", false
	case "npi":
		digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
		if len(digits) == 10 {
			return digits, true
		}
		return "", false
	case "imo":
		digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
		if len(digits) == 7 {
			return digits, true
		}
		return "", false
	case "qid":
		u := strings.ToUpper(strings.TrimSpace(s))
		if regexp.MustCompile(`^Q[1-9]\d*$`).MatchString(u) {
			return u, true
		}
		return "", false
	default:
		return s, true
	}
}
func (t *IdentifierType) Specificity(value string) float64 {
	n := len(value)
	if n <= 4 {
		return 0
	}
	if n >= 10 {
		return 1
	}
	return float64(n-4) / 6
}
func (t *IdentifierType) NodeID(value string) (string, bool)         { return "id:" + value, true }
func (t *IdentifierType) Caption(value string, format string) string { return value }
func (t *IdentifierType) Compare(left, right string) float64 {
    clean := func(s string) string { return strings.ToLower(regexp.MustCompile(`[\W_]+`).ReplaceAllString(s, "")) }
    l := clean(left)
    r := clean(right)
    if l == r { return 1.0 }
    if strings.Contains(l, r) || strings.Contains(r, l) {
        a, b := len(l), len(r)
        if a > b { a, b = b, a }
        if b == 0 { return 0 }
        return float64(a) / float64(b)
    }
    return 0.0
}

func normalizeIBAN(s string) string {
	s = strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	if !regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]{1,30}$`).MatchString(s) {
		return ""
	}
	rearranged := s[4:] + s[:4]
	num := big.NewInt(0)
	tmp := big.NewInt(0)
	ninetySeven := big.NewInt(97)
	for _, r := range rearranged {
		switch {
		case r >= '0' && r <= '9':
			tmp.SetInt64(int64(r - '0'))
			num.Mul(num, big.NewInt(10)).Add(num, tmp)
		case r >= 'A' && r <= 'Z':
			val := int64(int(r-'A') + 10)
			num.Mul(num, big.NewInt(100)).Add(num, big.NewInt(val))
		default:
			return ""
		}
		if num.BitLen() > 1200 {
			num.Mod(num, ninetySeven)
		}
	}
	num.Mod(num, ninetySeven)
	if num.Cmp(big.NewInt(1)) == 0 {
		return s
	}
	return ""
}
