package ftm

import (
    "net/url"
    "sort"
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

func normalizeURL(s string) (*url.URL, bool) {
    u, err := url.Parse(s)
    if err != nil || u.Scheme == "" {
        u, err = url.Parse("http://" + s)
        if err != nil { return nil, false }
    }
    u.Host = strings.ToLower(u.Host)
    // normalize query: sort parameters
    if u.RawQuery != "" {
        q := u.Query()
        keys := make([]string, 0, len(q))
        for k := range q { keys = append(keys, k) }
        sort.Strings(keys)
        nq := url.Values{}
        for _, k := range keys {
            vals := q[k]
            sort.Strings(vals)
            for _, v := range vals { nq.Add(k, v) }
        }
        u.RawQuery = nq.Encode()
    }
    u.Fragment = ""
    return u, true
}
func (t *URLType) Compare(left, right string) float64 {
    l, ok1 := normalizeURL(left)
    r, ok2 := normalizeURL(right)
    if !ok1 || !ok2 { return 0 }
    // Compare significant parts
    if l.Scheme == r.Scheme && l.Host == r.Host && strings.TrimSuffix(l.Path, "/") == strings.TrimSuffix(r.Path, "/") && l.RawQuery == r.RawQuery {
        return 1.0 * t.Specificity(l.String())
    }
    return 0.0
}
