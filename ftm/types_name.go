package ftm

import "strings"

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
