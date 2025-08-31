package ftm

import (
	"regexp"
	"strings"
)

var isoDateFull = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var isoDateMonth = regexp.MustCompile(`^\d{4}-\d{2}$`)
var isoDateYear = regexp.MustCompile(`^\d{4}$`)

// DateType supports YYYY, YYYY-MM, YYYY-MM-DD.
type DateType struct{ BaseType }

func NewDateType() *DateType {
	return &DateType{BaseType{name: "date", label: "Date", matchable: true}}
}
func (t *DateType) Validate(value string) bool {
	return isoDateFull.MatchString(value) || isoDateMonth.MatchString(value) || isoDateYear.MatchString(value)
}
func (t *DateType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`[^0-9-]`).ReplaceAllString(s, "")
	if t.Validate(s) {
		return s, true
	}
	return "", false
}
