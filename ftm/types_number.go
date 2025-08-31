package ftm

import (
	"strconv"
	"strings"
)

// NumberType stores numeric values.
type NumberType struct{ BaseType }

func NewNumberType() *NumberType {
	return &NumberType{BaseType{name: "number", label: "Number", matchable: true}}
}
func (t *NumberType) Validate(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}
func (t *NumberType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	if t.Validate(s) {
		return s, true
	}
	return "", false
}
