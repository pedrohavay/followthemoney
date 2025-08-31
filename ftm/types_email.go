package ftm

import (
	"golang.org/x/net/idna"
	"net/mail"
	"regexp"
	"strings"
)

// EmailType validates/normalizes emails with IDNA domain handling.
type EmailType struct{ BaseType }

func NewEmailType() *EmailType {
	return &EmailType{BaseType{name: "email", group: "emails", label: "E-Mail Address", matchable: true, pivot: true}}
}

var emailLocalRe = regexp.MustCompile(`^[^<>()[\]\\,;:\?\s@\"]{1,64}$`)
var domainLabelRe = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

func (t *EmailType) Validate(value string) bool { _, ok := t.Clean(value, false, "", nil); return ok }
func (t *EmailType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(strings.TrimPrefix(s, "mailto:"))
	if _, err := mail.ParseAddress(s); err == nil {
		if i := strings.LastIndex(s, "<"); i >= 0 {
			s = s[i+1:]
		}
		s = strings.TrimSuffix(s, ">")
	}
	at := strings.LastIndex(s, "@")
	if at <= 0 || at == len(s)-1 {
		return "", false
	}
	local, domain := s[:at], s[at+1:]
	if !emailLocalRe.MatchString(local) {
		return "", false
	}
	domain = strings.TrimSuffix(strings.ToLower(domain), ".")
	puny, err := idna.ToASCII(domain)
	if err != nil {
		return "", false
	}
	parts := strings.Split(puny, ".")
	if len(parts) < 2 {
		return "", false
	}
	for _, p := range parts {
		if !domainLabelRe.MatchString(p) {
			return "", false
		}
	}
	return local + "@" + strings.ToLower(puny), true
}
