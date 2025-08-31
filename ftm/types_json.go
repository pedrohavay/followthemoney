package ftm

import "encoding/json"

// JsonType packs/unpacks JSON values, not matchable.
type JsonType struct{ BaseType }

func NewJsonType() *JsonType {
	return &JsonType{BaseType{name: "json", label: "Nested data", matchable: false}}
}
func (t *JsonType) Validate(value string) bool {
	var v any
	return json.Unmarshal([]byte(value), &v) == nil
}
func (t *JsonType) Clean(raw string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	s, ok := sanitizeText(raw)
	if !ok {
		return "", false
	}
	var v any
	if json.Unmarshal([]byte(s), &v) == nil {
		return s, true
	}
	b, _ := json.Marshal(s)
	return string(b), true
}
func (t *JsonType) NodeID(string) (string, bool) { return "", false }
