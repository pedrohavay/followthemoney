package ftm

import (
	"regexp"
	"strings"
)

// ChecksumType assumes SHA1 hex (40 chars) by default.
type ChecksumType struct{ BaseType }

func NewChecksumType() *ChecksumType {
	return &ChecksumType{BaseType{name: "checksum", group: "checksums", label: "Checksum", matchable: true, pivot: true, maxLength: 40}}
}

var sha1Hex = regexp.MustCompile(`^[0-9a-f]{40}$`)

func (t *ChecksumType) Validate(value string) bool {
	return sha1Hex.MatchString(strings.ToLower(value))
}
func (t *ChecksumType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.ToLower(strings.TrimSpace(s))
	if sha1Hex.MatchString(s) {
		return s, true
	}
	return "", false
}
