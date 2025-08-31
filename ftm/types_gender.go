package ftm

import "strings"

// GenderType enum
type GenderType struct {
	BaseType
	values map[string]struct{}
}

func NewGenderType() *GenderType {
	return &GenderType{BaseType: BaseType{name: "gender", group: "genders", label: "Gender", matchable: false, maxLength: 16}, values: map[string]struct{}{"male": {}, "female": {}, "other": {}}}
}
func (t *GenderType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	code := strings.ToLower(strings.TrimSpace(text))
	switch code {
	case "m", "man", "masculin", "männlich", "мужской":
		code = "male"
	case "f", "woman", "féminin", "weiblich", "женский":
		code = "female"
	case "o", "d", "divers":
		code = "other"
	}
	if _, ok := t.values[code]; ok {
		return code, true
	}
	return "", false
}
func (t *GenderType) Validate(value string) bool { _, ok := t.values[value]; return ok }
