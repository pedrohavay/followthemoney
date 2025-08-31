package ftm

import (
    "regexp"
    "strings"
    levenshtein "github.com/agnivade/levenshtein"
)

// NameType simplifies personal/corporate names.
type NameType struct{ BaseType }

func NewNameType() *NameType {
	return &NameType{BaseType{name: "name", group: "names", label: "Name", matchable: true, pivot: true, maxLength: 512}}
}
func (t *NameType) Validate(value string) bool { return value != "" }
func (t *NameType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	// Strip quotes and collapse spaces
	text = strings.Trim(text, "\"' “”‘’")
	return sanitizeText(text)
}
func (t *NameType) Specificity(value string) float64 {
    n := float64(len(value))
    if n <= 3 {
        return 0
    }
    if n >= 50 {
        return 1
    }
    return (n - 3) / (50 - 3)
}

var nonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)
func normalizeNameTokens(s string) string {
    s = strings.ToLower(s)
    s = nonWord.ReplaceAllString(s, " ")
    s = strings.TrimSpace(s)
    for strings.Contains(s, "  ") { s = strings.ReplaceAll(s, "  ", " ") }
    return s
}
func similarity(a, b string) float64 {
    if len(a) == 0 || len(b) == 0 { return 0 }
    dist := levenshtein.ComputeDistance(a, b)
    maxlen := len(a)
    if len(b) > maxlen { maxlen = len(b) }
    if maxlen == 0 { return 0 }
    sim := 1.0 - float64(dist)/float64(maxlen)
    if sim < 0 { return 0 }
    return sim
}
func (t *NameType) Compare(left, right string) float64 {
    l := normalizeNameTokens(left)
    r := normalizeNameTokens(right)
    if l == "" || r == "" { return 0 }
    return similarity(l, r)
}
