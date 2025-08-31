package ftm

import (
	"regexp"
	"strings"
)

// CountryType accepts ISO-3166 alpha-2 codes and common FtM codes.
type CountryType struct{ BaseType }

func NewCountryType() *CountryType {
	return &CountryType{BaseType{name: "country", group: "countries", label: "Country", matchable: true, maxLength: 16}}
}

var countryAlpha2 = regexp.MustCompile(`^[A-Za-z]{2}$`)
var ftmCountryCodes = make(map[string]struct{})

func init() {
	codes := []string{"ae", "af", "al", "am", "ao", "ar", "at", "au", "az", "ba", "bd", "be", "bg", "bh", "bi", "bj", "bo", "br", "bs", "bw", "by", "bz", "ca", "cd", "cf", "cg", "ch", "ci", "cl", "cm", "cn", "co", "cr", "cu", "cz", "de", "dk", "do", "dz", "ec", "ee", "eg", "er", "es", "et", "fi", "fj", "fr", "ga", "gb", "ge", "gh", "gm", "gn", "gq", "gr", "gt", "gw", "gy", "hk", "hn", "hr", "ht", "hu", "id", "ie", "il", "in", "iq", "ir", "is", "it", "jm", "jo", "jp", "ke", "kg", "kh", "km", "kp", "kr", "kw", "kz", "la", "lb", "lk", "lr", "ls", "lt", "lu", "lv", "ly", "ma", "md", "me", "mg", "mk", "ml", "mm", "mn", "mr", "mt", "mu", "mw", "mx", "my", "mz", "na", "ne", "ng", "ni", "nl", "no", "np", "nz", "om", "pa", "pe", "pg", "ph", "pk", "pl", "ps", "pt", "py", "qa", "ro", "rs", "ru", "rw", "sa", "sd", "se", "sg", "si", "sk", "sl", "sn", "so", "ss", "sv", "sy", "sz", "td", "tg", "th", "tj", "tl", "tm", "tn", "tr", "tt", "tw", "tz", "ua", "ug", "us", "uy", "uz", "ve", "vn", "ye", "za", "zm", "zw"}
	for _, c := range codes {
		ftmCountryCodes[c] = struct{}{}
	}
}

func (t *CountryType) Validate(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	if _, ok := ftmCountryCodes[v]; ok {
		return true
	}
	return countryAlpha2.MatchString(value)
}
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
