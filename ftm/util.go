package ftm

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strings"
	"unicode"
)

// sanitizeText normalizes input to NFC-ish ASCII-safe representation.
// It trims spaces, removes control characters, and collapses internal whitespace.
func sanitizeText(s string) (string, bool) {
	if s == "" {
		return "", false
	}
	// Remove control characters and normalize spaces
	b := strings.Builder{}
	lastSpace := false
	for _, r := range s {
		if r == '\u0000' {
			continue
		}
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			continue
		}
		if unicode.IsSpace(r) {
			if lastSpace {
				continue
			}
			b.WriteRune(' ')
			lastSpace = true
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "", false
	}
	// Enforce a sensible max length to avoid pathological inputs
	if len(out) > 10000 {
		out = out[:10000]
	}
	return out, true
}

// joinText joins non-empty parts with the given separator.
func joinText(sep string, parts ...string) (string, bool) {
	xs := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		xs = append(xs, p)
	}
	if len(xs) == 0 {
		return "", false
	}
	return strings.Join(xs, sep), true
}

// makeEntityID hashes the provided parts with an optional key prefix.
func makeEntityID(keyPrefix string, parts ...string) (string, bool) {
	h := sha1.New()
	if keyPrefix != "" {
		h.Write([]byte(keyPrefix))
	}
	base := h.Sum(nil)
	h.Reset()
	if keyPrefix != "" {
		h.Write([]byte(keyPrefix))
	}
	for _, p := range parts {
		if p != "" {
			h.Write([]byte(p))
		}
	}
	out := h.Sum(nil)
	if string(out) == string(base) {
		return "", false
	}
	return hex.EncodeToString(out), true
}

// shortest returns the shortest non-empty string.
func shortest(values ...string) string {
	nonEmpty := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			nonEmpty = append(nonEmpty, v)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	sort.Slice(nonEmpty, func(i, j int) bool { return len(nonEmpty[i]) < len(nonEmpty[j]) })
	return nonEmpty[0]
}

// date regexes moved to types_date.go
