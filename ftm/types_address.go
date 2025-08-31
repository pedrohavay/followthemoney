package ftm

import (
	"regexp"
	"strings"
)

// AddressType normalizes lines/commas and collapses spaces.
type AddressType struct{ BaseType }

func NewAddressType() *AddressType {
	return &AddressType{BaseType{name: "address", group: "addresses", label: "Address", matchable: true, pivot: true}}
}

var addrLineBreaks = regexp.MustCompile(`(\r\n|\n|<BR/>|<BR>|\t|ESQ\.,|ESQ,|;)`)
var addrCommata = regexp.MustCompile(`(,\s?[,\.])`)

func (t *AddressType) Validate(value string) bool { _, ok := t.Clean(value, false, "", nil); return ok }
func (t *AddressType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = addrLineBreaks.ReplaceAllString(s, ", ")
	s = addrCommata.ReplaceAllString(s, ", ")
	s = strings.TrimSpace(s)
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	if s == "" {
		return "", false
	}
	return s, true
}
func (t *AddressType) NodeID(value string) (string, bool) {
	v, ok := sanitizeText(strings.ToLower(value))
	if !ok {
		return "", false
	}
	v = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(v, "-")
	v = strings.Trim(v, "-")
	if v == "" {
		return "", false
	}
	return "addr:" + v, true
}
