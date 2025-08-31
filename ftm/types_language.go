package ftm

import "strings"

// LanguageType with ISO-639-3 whitelist.
type LanguageType struct{ BaseType }

func NewLanguageType() *LanguageType {
	return &LanguageType{BaseType{name: "language", group: "languages", label: "Language", matchable: false, maxLength: 16}}
}

var languageWhitelist = map[string]struct{}{}

func init() {
	for _, l := range []string{"eng", "fra", "deu", "rus", "spa", "nld", "ron", "kat", "ara", "tur", "ltz", "ell", "lit", "ukr", "zho", "bel", "bul", "bos", "jpn", "ces", "lav", "por", "pol", "hye", "hrv", "hin", "heb", "uzb", "mon", "urd", "sqi", "kor", "isl", "ita", "est", "nor", "fas", "swa", "slv", "slk", "aze", "tgk", "kaz", "tuk", "kir", "hun", "dan", "afr", "swe", "srp", "ind", "kan", "mkd", "mlt", "msa", "fin", "cat", "nep", "tgl", "fil", "mya", "khm", "cnr", "ben"} {
		languageWhitelist[l] = struct{}{}
	}
}

func (t *LanguageType) Validate(value string) bool {
	_, ok := languageWhitelist[strings.ToLower(strings.TrimSpace(value))]
	return ok
}
func (t *LanguageType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	code := strings.ToLower(strings.TrimSpace(text))
	if _, ok := languageWhitelist[code]; ok {
		return code, true
	}
	return "", false
}
