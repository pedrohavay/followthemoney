package ftm

// TextType is for long text; not matchable.
type TextType struct{ BaseType }

func NewTextType() *TextType {
	return &TextType{BaseType{name: "text", label: "Text", matchable: false, maxLength: 65000, totalSize: 30 * 1024 * 1024}}
}
func (t *TextType) Validate(value string) bool { return value != "" }
func (t *TextType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	return sanitizeText(text)
}
