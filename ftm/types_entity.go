package ftm

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
