package ftm

import (
	"strings"

	phonenumbers "github.com/nyaruka/phonenumbers"
)

// PhoneType uses libphonenumber for E.164 formatting and parsing.
type PhoneType struct{ BaseType }

func NewPhoneType() *PhoneType {
	return &PhoneType{BaseType{name: "phone", group: "phones", label: "Phone number", matchable: true, pivot: true, maxLength: 64}}
}
func (t *PhoneType) Validate(value string) bool {
	n, err := phonenumbers.Parse(value, "")
	if err != nil {
		return false
	}
	return phonenumbers.IsValidNumber(n)
}
func (t *PhoneType) Clean(text string, _ bool, _ string, proxy *EntityProxy) (string, bool) {
	s, ok := sanitizeText(text)
	if !ok {
		return "", false
	}
	regions := []string{""}
	if proxy != nil {
		cc := proxy.Countries()
		tmp := make([]string, 0, len(cc))
		for _, c := range cc {
			if len(c) == 2 {
				tmp = append(tmp, strings.ToUpper(c))
			}
		}
		if len(tmp) > 0 {
			regions = append(tmp, regions...)
		}
	}
	for _, region := range regions {
		n, err := phonenumbers.Parse(s, region)
		if err != nil {
			continue
		}
		if phonenumbers.IsValidNumber(n) {
			return phonenumbers.Format(n, phonenumbers.E164), true
		}
	}
	return "", false
}
func (t *PhoneType) CountryHint(value string) (string, bool) {
	n, err := phonenumbers.Parse(value, "")
	if err != nil {
		return "", false
	}
	reg := phonenumbers.GetRegionCodeForNumber(n)
	if reg == "" {
		return "", false
	}
	return strings.ToLower(reg), true
}
func (t *PhoneType) NodeID(value string) (string, bool) { return "tel:" + value, true }
