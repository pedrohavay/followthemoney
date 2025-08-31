package ftm

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// PropertyType defines behavior for a value type.
// Implementations should be stateless and reusable.
type PropertyType interface {
	Name() string
	Group() string   // logical group name, may be empty
	Label() string   // human-readable name
	Matchable() bool // included in matching/comparison
	Pivot() bool     // used to form graph pivots
	MaxLength() int  // maximum length of a single value
	TotalSize() int  // limit to total accumulated string length (0 = unlimited)

	Validate(value string) bool
	Clean(text string, fuzzy bool, format string, proxy *EntityProxy) (string, bool)
	Specificity(value string) float64
	Caption(value string, format string) string
	NodeID(value string) (string, bool)
	CountryHint(value string) (string, bool)
}

// BaseType offers default implementations.
type BaseType struct {
	name      string
	group     string
	label     string
	matchable bool
	pivot     bool
	maxLength int
	totalSize int
}

func (b BaseType) Name() string                          { return b.name }
func (b BaseType) Group() string                         { return b.group }
func (b BaseType) Label() string                         { return b.label }
func (b BaseType) Matchable() bool                       { return b.matchable }
func (b BaseType) Pivot() bool                           { return b.pivot }
func (b BaseType) MaxLength() int                        { return b.maxLength }
func (b BaseType) TotalSize() int                        { return b.totalSize }
func (b BaseType) Specificity(string) float64            { return 0.0 }
func (b BaseType) Caption(value string, _ string) string { return value }
func (b BaseType) NodeID(value string) (string, bool) {
	// Default node id: prefix:name:<slug>
	s, ok := sanitizeText(strings.ToLower(value))
	if !ok {
		return "", false
	}
	s = strings.ReplaceAll(s, " ", "-")
	s = regexp.MustCompile(`[^a-z0-9._-]`).ReplaceAllString(s, "")
	if s == "" {
		return "", false
	}
	return b.name + ":" + s, true
}
func (b BaseType) CountryHint(string) (string, bool) { return "", false }

// StringType is the catch-all for most text.
type StringType struct{ BaseType }

func NewStringType() *StringType {
	return &StringType{BaseType{name: "string", label: "String", matchable: true, maxLength: 2048}}
}
func (t *StringType) Validate(value string) bool { _, ok := sanitizeText(value); return ok }
func (t *StringType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	return sanitizeText(text)
}

// TextType is for long text; not matchable.
type TextType struct{ BaseType }

func NewTextType() *TextType {
	return &TextType{BaseType{name: "text", label: "Text", matchable: false, maxLength: 100000}}
}
func (t *TextType) Validate(value string) bool { return value != "" }
func (t *TextType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	return sanitizeText(text)
}

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
	// Simple heuristic: longer names are more specific up to a cap.
	n := float64(len(value))
	if n <= 3 {
		return 0
	}
	if n >= 50 {
		return 1
	}
	return (n - 3) / (50 - 3)
}

// DateType supports YYYY, YYYY-MM, YYYY-MM-DD.
type DateType struct{ BaseType }

func NewDateType() *DateType {
	return &DateType{BaseType{name: "date", label: "Date", matchable: true}}
}
func (t *DateType) Validate(value string) bool {
	if isoDateFull.MatchString(value) || isoDateMonth.MatchString(value) || isoDateYear.MatchString(value) {
		return true
	}
	return false
}
func (t *DateType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	// Allow only digits and '-'
	s = regexp.MustCompile(`[^0-9-]`).ReplaceAllString(s, "")
	if t.Validate(s) {
		return s, true
	}
	return "", false
}

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

// URLType validates URLs.
type URLType struct{ BaseType }

func NewURLType() *URLType { return &URLType{BaseType{name: "url", label: "URL", matchable: true}} }
func (t *URLType) Validate(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme != "" && u.Host != ""
}
func (t *URLType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	if !t.Validate(s) {
		return "", false
	}
	return s, true
}
func (t *URLType) NodeID(value string) (string, bool) { return "url:" + value, true }

// CountryType accepts ISO-3166 alpha-2 codes, lowercased.
type CountryType struct{ BaseType }

func NewCountryType() *CountryType {
	return &CountryType{BaseType{name: "country", group: "countries", label: "Country", matchable: true, maxLength: 2}}
}

var countryAlpha2 = regexp.MustCompile(`^[A-Za-z]{2}$`)

func (t *CountryType) Validate(value string) bool { return countryAlpha2.MatchString(value) }
func (t *CountryType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	s = strings.ToLower(strings.TrimSpace(s))
	if t.Validate(s) {
		return s, true
	}
	return "", false
}

// EntityType references another entity by id.
type EntityType struct{ BaseType }

func NewEntityType() *EntityType {
	return &EntityType{BaseType{name: "entity", label: "Entity", matchable: false}}
}
func (t *EntityType) Validate(value string) bool { return value != "" }
func (t *EntityType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	return sanitizeText(text)
}
func (t *EntityType) NodeID(value string) (string, bool) { return value, value != "" }

// Registry holds known property types and helpers.
type Registry struct {
	// Commonly referenced types
	String  *StringType
	Text    *TextType
	Name    *NameType
	Date    *DateType
	Number  *NumberType
	URL     *URLType
	Country *CountryType
	Entity  *EntityType

	types     map[string]PropertyType
	matchable map[string]PropertyType
	pivots    map[string]PropertyType
	groups    map[string]PropertyType
}

func NewRegistry() *Registry {
	r := &Registry{
		String:    NewStringType(),
		Text:      NewTextType(),
		Name:      NewNameType(),
		Date:      NewDateType(),
		Number:    NewNumberType(),
		URL:       NewURLType(),
		Country:   NewCountryType(),
		Entity:    NewEntityType(),
		types:     map[string]PropertyType{},
		matchable: map[string]PropertyType{},
		pivots:    map[string]PropertyType{},
		groups:    map[string]PropertyType{},
	}
	for _, t := range []PropertyType{r.String, r.Text, r.Name, r.Date, r.Number, r.URL, r.Country, r.Entity} {
		r.types[t.Name()] = t
		if t.Matchable() {
			r.matchable[t.Name()] = t
		}
		if t.Pivot() {
			r.pivots[t.Name()] = t
		}
		if g := t.Group(); g != "" {
			r.groups[g] = t
		}
	}
	return r
}

func (r *Registry) Get(name string) PropertyType { return r.types[name] }

var registry = NewRegistry()
