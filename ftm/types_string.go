package ftm

// StringType is the catch-all for most text.
type StringType struct{ BaseType }

func NewStringType() *StringType {
	return &StringType{BaseType{name: "string", label: "String", matchable: false, maxLength: 1024}}
}
func (t *StringType) Validate(value string) bool { _, ok := sanitizeText(value); return ok }
func (t *StringType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	return sanitizeText(text)
}
