package ftm

// HTMLType mirrors TextType but signals HTML content.
type HTMLType struct{ BaseType }

func NewHTMLType() *HTMLType {
	return &HTMLType{BaseType{name: "html", label: "HTML", matchable: false, maxLength: 65000, totalSize: 30 * 1024 * 1024}}
}
func (t *HTMLType) Validate(value string) bool { return value != "" }
func (t *HTMLType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	return sanitizeText(text)
}
