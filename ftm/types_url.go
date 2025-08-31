package ftm

import (
	"net/url"
	"strings"
)

// URLType validates URLs and normalizes.
type URLType struct{ BaseType }

func NewURLType() *URLType {
	return &URLType{BaseType{name: "url", label: "URL", matchable: true, maxLength: 4096}}
}
func (t *URLType) Validate(value string) bool {
	u, err := url.Parse(value)
	if err != nil {
		return false
	}
	if u.Scheme == "" {
		u, err = url.Parse("http://" + value)
		if err != nil {
			return false
		}
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "ftp", "mailto":
		return u.Host != "" || u.Scheme == "mailto"
	default:
		return false
	}
}
func (t *URLType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" {
		u, err = url.Parse("http://" + s)
	}
	if err != nil || !t.Validate(u.String()) {
		return "", false
	}
	u.Host = strings.ToLower(u.Host)
	return u.String(), true
}
func (t *URLType) NodeID(value string) (string, bool) { return "url:" + value, true }
