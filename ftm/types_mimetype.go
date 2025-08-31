package ftm

import (
	"regexp"
	"strings"
)

// MimeType validates type/subtype tokens.
type MimeType struct{ BaseType }

func NewMimeType() *MimeType {
	return &MimeType{BaseType{name: "mimetype", group: "mimetypes", label: "MIME-Type", matchable: false}}
}

var mimeRe = regexp.MustCompile(`^[a-zA-Z0-9!#$&^_.+-]{1,127}/[a-zA-Z0-9!#$&^_.+-]{1,127}$`)

func (t *MimeType) Validate(value string) bool { return mimeRe.MatchString(value) }
func (t *MimeType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" || s == "application/octet-stream" {
		return "", false
	}
	if mimeRe.MatchString(s) {
		return s, true
	}
	return "", false
}
