package ftm

import (
    "encoding/json"
    "math/big"
    "net"
    "net/mail"
    "net/url"
    "regexp"
    "strconv"
    "strings"

    phonenumbers "github.com/nyaruka/phonenumbers"
    "golang.org/x/net/idna"
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

func NewStringType() *StringType { return &StringType{BaseType{name: "string", label: "String", matchable: false, maxLength: 1024}} }
func (t *StringType) Validate(value string) bool { _, ok := sanitizeText(value); return ok }
func (t *StringType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
	return sanitizeText(text)
}

// TextType is for long text; not matchable.
type TextType struct{ BaseType }

func NewTextType() *TextType { return &TextType{BaseType{name: "text", label: "Text", matchable: false, maxLength: 65000, totalSize: 30 * 1024 * 1024}} }
func (t *TextType) Validate(value string) bool { return value != "" }
func (t *TextType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
    return sanitizeText(text)
}

// HTMLType mirrors TextType but signals HTML content.
type HTMLType struct{ BaseType }

func NewHTMLType() *HTMLType {
    return &HTMLType{BaseType{name: "html", label: "HTML", matchable: false, maxLength: 65000, totalSize: 30 * 1024 * 1024}}
}
func (t *HTMLType) Validate(value string) bool { return value != "" }
func (t *HTMLType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) { return sanitizeText(text) }

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

func NewURLType() *URLType { return &URLType{BaseType{name: "url", label: "URL", matchable: true, maxLength: 4096}} }
func (t *URLType) Validate(value string) bool {
    u, err := url.Parse(value)
    if err != nil { return false }
    if u.Scheme == "" {
        u, err = url.Parse("http://" + value)
        if err != nil { return false }
    }
    switch strings.ToLower(u.Scheme) {
    case "http", "https", "ftp", "mailto":
        return u.Host != "" || u.Scheme == "mailto"
    default:
        return false
    }
}
func (t *URLType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
    s, ok := sanitizeText(text); if !ok { return "", false }
    s = strings.TrimSpace(s)
    u, err := url.Parse(s)
    if err != nil || u.Scheme == "" { u, err = url.Parse("http://" + s) }
    if err != nil || !t.Validate(u.String()) { return "", false }
    u.Host = strings.ToLower(u.Host)
    return u.String(), true
}
func (t *URLType) NodeID(value string) (string, bool) { return "url:" + value, true }

// CountryType accepts ISO-3166 alpha-2 codes, lowercased.
type CountryType struct{ BaseType }

func NewCountryType() *CountryType { return &CountryType{BaseType{name: "country", group: "countries", label: "Country", matchable: true, maxLength: 16}} }
var countryAlpha2 = regexp.MustCompile(`^[A-Za-z]{2}$`)
var ftmCountryCodes = make(map[string]struct{})
func init() {
    codes := []string{"ae","af","al","am","ao","ar","at","au","az","ba","bd","be","bg","bh","bi","bj","bo","br","bs","bw","by","bz","ca","cd","cf","cg","ch","ci","cl","cm","cn","co","cr","cu","cz","de","dk","do","dz","ec","ee","eg","er","es","et","fi","fj","fr","ga","gb","ge","gh","gm","gn","gq","gr","gt","gw","gy","hk","hn","hr","ht","hu","id","ie","il","in","iq","ir","is","it","jm","jo","jp","ke","kg","kh","km","kp","kr","kw","kz","la","lb","lk","lr","ls","lt","lu","lv","ly","ma","md","me","mg","mk","ml","mm","mn","mr","mt","mu","mw","mx","my","mz","na","ne","ng","ni","nl","no","np","nz","om","pa","pe","pg","ph","pk","pl","ps","pt","py","qa","ro","rs","ru","rw","sa","sd","se","sg","si","sk","sl","sn","so","ss","sv","sy","sz","td","tg","th","tj","tl","tm","tn","tr","tt","tw","tz","ua","ug","us","uy","uz","ve","vn","ye","za","zm","zw"}
    for _, c := range codes { ftmCountryCodes[c] = struct{}{} }
}
func (t *CountryType) Validate(value string) bool {
    v := strings.ToLower(strings.TrimSpace(value))
    if _, ok := ftmCountryCodes[v]; ok { return true }
    return countryAlpha2.MatchString(value)
}
func (t *CountryType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
    s, ok := sanitizeText(text); if !ok { return "", false }
    s = strings.ToLower(strings.TrimSpace(s))
    if t.Validate(s) { return s, true }
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

// EmailType validates/normalizes emails with IDNA domain handling.
type EmailType struct{ BaseType }
func NewEmailType() *EmailType { return &EmailType{BaseType{name: "email", group: "emails", label: "E-Mail Address", matchable: true, pivot: true}} }
var emailLocalRe = regexp.MustCompile(`^[^<>()[\]\\,;:\?\s@\"]{1,64}$`)
var domainLabelRe = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)
func (t *EmailType) Validate(value string) bool { _, ok := t.Clean(value, false, "", nil); return ok }
func (t *EmailType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
    s, ok := sanitizeText(text); if !ok { return "", false }
    s = strings.TrimSpace(strings.TrimPrefix(s, "mailto:"))
    if _, err := mail.ParseAddress(s); err == nil {
        if i := strings.LastIndex(s, "<"); i >= 0 { s = s[i+1:] }
        s = strings.TrimSuffix(s, ">")
    }
    at := strings.LastIndex(s, "@")
    if at <= 0 || at == len(s)-1 { return "", false }
    local, domain := s[:at], s[at+1:]
    if !emailLocalRe.MatchString(local) { return "", false }
    domain = strings.TrimSuffix(strings.ToLower(domain), ".")
    puny, err := idna.ToASCII(domain)
    if err != nil { return "", false }
    parts := strings.Split(puny, ".")
    if len(parts) < 2 { return "", false }
    for _, p := range parts { if !domainLabelRe.MatchString(p) { return "", false } }
    return local + "@" + strings.ToLower(puny), true
}

// IpType validates IPv4/IPv6.
type IpType struct{ BaseType }
func NewIpType() *IpType { return &IpType{BaseType{name: "ip", group: "ips", label: "IP Address", matchable: true, pivot: true, maxLength: 64}} }
func (t *IpType) Validate(value string) bool { return net.ParseIP(value) != nil }
func (t *IpType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) { s, ok := sanitizeText(text); if !ok { return "", false }; ip := net.ParseIP(strings.TrimSpace(s)); if ip == nil { return "", false }; return ip.String(), true }

// PhoneType uses libphonenumber for E.164 formatting and parsing.
type PhoneType struct{ BaseType }
func NewPhoneType() *PhoneType { return &PhoneType{BaseType{name: "phone", group: "phones", label: "Phone number", matchable: true, pivot: true, maxLength: 64}} }
func (t *PhoneType) Validate(value string) bool { n, err := phonenumbers.Parse(value, ""); if err != nil { return false }; return phonenumbers.IsValidNumber(n) }
func (t *PhoneType) Clean(text string, _ bool, _ string, proxy *EntityProxy) (string, bool) {
    s, ok := sanitizeText(text); if !ok { return "", false }
    regions := []string{""}
    if proxy != nil {
        cc := proxy.Countries()
        tmp := make([]string, 0, len(cc))
        for _, c := range cc { if len(c) == 2 { tmp = append(tmp, strings.ToUpper(c)) } }
        if len(tmp) > 0 { regions = append(tmp, regions...) }
    }
    for _, region := range regions {
        n, err := phonenumbers.Parse(s, region)
        if err != nil { continue }
        if phonenumbers.IsValidNumber(n) { return phonenumbers.Format(n, phonenumbers.E164), true }
    }
    return "", false
}
func (t *PhoneType) CountryHint(value string) (string, bool) { n, err := phonenumbers.Parse(value, ""); if err != nil { return "", false }; reg := phonenumbers.GetRegionCodeForNumber(n); if reg == "" { return "", false }; return strings.ToLower(reg), true }
func (t *PhoneType) NodeID(value string) (string, bool) { return "tel:" + value, true }

// AddressType normalizes lines/commas and collapses spaces.
type AddressType struct{ BaseType }
func NewAddressType() *AddressType { return &AddressType{BaseType{name: "address", group: "addresses", label: "Address", matchable: true, pivot: true}} }
var addrLineBreaks = regexp.MustCompile(`(\r\n|\n|<BR/>|<BR>|\t|ESQ\.,|ESQ,|;)`)
var addrCommata = regexp.MustCompile(`(,\s?[,\.])`)
func (t *AddressType) Validate(value string) bool { _, ok := t.Clean(value, false, "", nil); return ok }
func (t *AddressType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
    s, ok := sanitizeText(text); if !ok { return "", false }
    s = addrLineBreaks.ReplaceAllString(s, ", ")
    s = addrCommata.ReplaceAllString(s, ", ")
    s = strings.TrimSpace(s)
    for strings.Contains(s, "  ") { s = strings.ReplaceAll(s, "  ", " ") }
    if s == "" { return "", false }
    return s, true
}
func (t *AddressType) NodeID(value string) (string, bool) { v, ok := sanitizeText(strings.ToLower(value)); if !ok { return "", false }; v = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(v, "-"); v = strings.Trim(v, "-"); if v == "" { return "", false }; return "addr:" + v, true }

// LanguageType with ISO-639-3 whitelist.
type LanguageType struct{ BaseType }
func NewLanguageType() *LanguageType { return &LanguageType{BaseType{name: "language", group: "languages", label: "Language", matchable: false, maxLength: 16}} }
var languageWhitelist = map[string]struct{}{}
func init() { for _, l := range []string{"eng","fra","deu","rus","spa","nld","ron","kat","ara","tur","ltz","ell","lit","ukr","zho","bel","bul","bos","jpn","ces","lav","por","pol","hye","hrv","hin","heb","uzb","mon","urd","sqi","kor","isl","ita","est","nor","fas","swa","slv","slk","aze","tgk","kaz","tuk","kir","hun","dan","afr","swe","srp","ind","kan","mkd","mlt","msa","fin","cat","nep","tgl","fil","mya","khm","cnr","ben"} { languageWhitelist[l] = struct{}{} } }
func (t *LanguageType) Validate(value string) bool { _, ok := languageWhitelist[strings.ToLower(strings.TrimSpace(value))]; return ok }
func (t *LanguageType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) { code := strings.ToLower(strings.TrimSpace(text)); if _, ok := languageWhitelist[code]; ok { return code, true }; return "", false }

// MimeType validates type/subtype tokens.
type MimeType struct{ BaseType }
func NewMimeType() *MimeType { return &MimeType{BaseType{name: "mimetype", group: "mimetypes", label: "MIME-Type", matchable: false}} }
var mimeRe = regexp.MustCompile(`^[a-zA-Z0-9!#$&^_.+-]{1,127}/[a-zA-Z0-9!#$&^_.+-]{1,127}$`)
func (t *MimeType) Validate(value string) bool { return mimeRe.MatchString(value) }
func (t *MimeType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) { s, ok := sanitizeText(text); if !ok { return "", false }; s = strings.ToLower(strings.TrimSpace(s)); if s == "" || s == "application/octet-stream" { return "", false }; if mimeRe.MatchString(s) { return s, true }; return "", false }

// ChecksumType assumes SHA1 hex (40 chars) by default.
type ChecksumType struct{ BaseType }
func NewChecksumType() *ChecksumType { return &ChecksumType{BaseType{name: "checksum", group: "checksums", label: "Checksum", matchable: true, pivot: true, maxLength: 40}} }
var sha1Hex = regexp.MustCompile(`^[0-9a-f]{40}$`)
func (t *ChecksumType) Validate(value string) bool { return sha1Hex.MatchString(strings.ToLower(value)) }
func (t *ChecksumType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) { s, ok := sanitizeText(text); if !ok { return "", false }; s = strings.ToLower(strings.TrimSpace(s)); if sha1Hex.MatchString(s) { return s, true }; return "", false }

// IdentifierType with optional format validation (IBAN, LEI).
type IdentifierType struct{ BaseType }
func NewIdentifierType() *IdentifierType { return &IdentifierType{BaseType{name: "identifier", group: "identifiers", label: "Identifier", matchable: true, pivot: true, maxLength: 64}} }
var identCleanRe = regexp.MustCompile(`[\W_]+`)
func (t *IdentifierType) Validate(value string) bool { _, ok := t.Clean(value, false, "", nil); return ok }
func (t *IdentifierType) Clean(text string, _ bool, format string, _ *EntityProxy) (string, bool) {
    s, ok := sanitizeText(text); if !ok { return "", false }
    s = strings.TrimSpace(s)
    switch strings.ToLower(format) {
    case "iban":
        if iban := normalizeIBAN(s); iban != "" { return iban, true }
        return "", false
    case "lei":
        u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
        if regexp.MustCompile(`^[A-Z0-9]{20}$`).MatchString(u) { return u, true }
        return "", false
    case "bic":
        u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
        if regexp.MustCompile(`^[A-Z]{4}[A-Z]{2}[A-Z0-9]{2}([A-Z0-9]{3})?$`).MatchString(u) { return u, true }
        return "", false
    case "isin":
        u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
        if regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`).MatchString(u) { return u, true }
        return "", false
    case "figi":
        u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
        if regexp.MustCompile(`^[A-Z0-9]{12}$`).MatchString(u) { return u, true }
        return "", false
    case "ssn":
        digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
        if len(digits) == 9 { return digits, true }
        return "", false
    case "uscc":
        u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
        if regexp.MustCompile(`^[0-9A-Z]{18}$`).MatchString(u) { return u, true }
        return "", false
    case "inn":
        digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
        if len(digits) == 10 || len(digits) == 12 { return digits, true }
        return "", false
    case "ogrn":
        digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
        if len(digits) == 13 || len(digits) == 15 { return digits, true }
        return "", false
    case "uei":
        u := strings.ToUpper(strings.ReplaceAll(s, " ", ""))
        if regexp.MustCompile(`^[A-Z0-9]{12}$`).MatchString(u) { return u, true }
        return "", false
    case "npi":
        digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
        if len(digits) == 10 { return digits, true }
        return "", false
    case "imo":
        digits := regexp.MustCompile(`\D`).ReplaceAllString(s, "")
        if len(digits) == 7 { return digits, true }
        return "", false
    case "qid":
        u := strings.ToUpper(strings.TrimSpace(s))
        if regexp.MustCompile(`^Q[1-9]\d*$`).MatchString(u) { return u, true }
        return "", false
    default:
        return s, true
    }
}
func (t *IdentifierType) Specificity(value string) float64 { n := len(value); if n <= 4 { return 0 }; if n >= 10 { return 1 }; return float64(n-4)/6 }
func (t *IdentifierType) NodeID(value string) (string, bool) { return "id:" + value, true }
func (t *IdentifierType) Caption(value string, format string) string { return value }

func normalizeIBAN(s string) string {
    s = strings.ToUpper(strings.ReplaceAll(s, " ", ""))
    if !regexp.MustCompile(`^[A-Z]{2}[0-9]{2}[A-Z0-9]{1,30}$`).MatchString(s) { return "" }
    rearranged := s[4:] + s[:4]
    num := big.NewInt(0)
    tmp := big.NewInt(0)
    ninetySeven := big.NewInt(97)
    for _, r := range rearranged {
        switch {
        case r >= '0' && r <= '9':
            tmp.SetInt64(int64(r - '0'))
            num.Mul(num, big.NewInt(10)).Add(num, tmp)
        case r >= 'A' && r <= 'Z':
            val := int64(int(r-'A') + 10)
            num.Mul(num, big.NewInt(100)).Add(num, big.NewInt(val))
        default:
            return ""
        }
        if num.BitLen() > 1200 { num.Mod(num, ninetySeven) }
    }
    num.Mod(num, ninetySeven)
    if num.Cmp(big.NewInt(1)) == 0 { return s }
    return ""
}

// TopicType as enum with fixed set.
type TopicType struct{ BaseType; values map[string]string }
func NewTopicType() *TopicType {
    t := &TopicType{BaseType: BaseType{name: "topic", group: "topics", label: "Topic", matchable: false, maxLength: 64}, values: map[string]string{}}
    for k, v := range map[string]string{
        "crime":"Crime","crime.fraud":"Fraud","crime.cyber":"Cybercrime","crime.fin":"Financial crime","crime.env":"Environmental violations","crime.theft":"Theft","crime.war":"War crimes","crime.boss":"Criminal leadership","crime.terror":"Terrorism","crime.traffick":"Trafficking","crime.traffick.drug":"Drug trafficking","crime.traffick.human":"Human trafficking","forced.labor":"Forced labor","asset.frozen":"Frozen asset","wanted":"Wanted","corp.offshore":"Offshore","corp.shell":"Shell company","corp.public":"Public listed company","corp.disqual":"Disqualified","gov":"Government","gov.national":"National government","gov.state":"State government","gov.muni":"Municipal government","gov.soe":"State-owned enterprise","gov.igo":"Intergovernmental organization","gov.head":"Head of government or state","gov.admin":"Civil service","gov.executive":"Executive branch of government","gov.legislative":"Legislative branch of government","gov.judicial":"Judicial branch of government","gov.security":"Security services","gov.financial":"Central banking and financial integrity","gov.religion":"Religious leadership","fin":"Financial services","fin.bank":"Bank","fin.fund":"Fund","fin.adivsor":"Financial advisor","mare.detained":"Maritime detention","mare.shadow":"Shadow fleet","mare.sts":"Ship-to-ship transfer","reg.action":"Regulator action","reg.warn":"Regulator warning","role.pep":"Politician","role.pol":"Non-PEP","role.rca":"Close Associate","role.judge":"Judge","role.civil":"Civil servant","role.diplo":"Diplomat","role.lawyer":"Lawyer","role.acct":"Accountant","role.spy":"Spy","role.oligarch":"Oligarch","role.journo":"Journalist","role.act":"Activist","role.lobby":"Lobbyist","pol.party":"Political party","pol.union":"Union","rel":"Religion","mil":"Military","sanction":"Sanctioned entity","sanction.linked":"Sanction-linked entity","sanction.counter":"Counter-sanctioned entity","export.control":"Export controlled","export.risk":"Trade risk","debarment":"Debarred entity","poi":"Person of interest",
    } { t.values[k] = v }
    return t
}
func (t *TopicType) Validate(value string) bool { _, ok := t.values[strings.ToLower(strings.TrimSpace(value))]; return ok }
func (t *TopicType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) { v := strings.ToLower(strings.TrimSpace(text)); if _, ok := t.values[v]; ok { return v, true }; return "", false }
func (t *TopicType) Caption(value, _ string) string { if l, ok := t.values[value]; ok { return l }; return value }

// GenderType enum
type GenderType struct{ BaseType; values map[string]struct{} }
func NewGenderType() *GenderType { return &GenderType{BaseType: BaseType{name: "gender", group: "genders", label: "Gender", matchable: false, maxLength: 16}, values: map[string]struct{}{"male":{},"female":{},"other":{}}} }
func (t *GenderType) Clean(text string, _ bool, _ string, _ *EntityProxy) (string, bool) {
    code := strings.ToLower(strings.TrimSpace(text))
    switch code {
    case "m","man","masculin","männlich","мужской": code = "male"
    case "f","woman","féminin","weiblich","женский": code = "female"
    case "o","d","divers": code = "other"
    }
    if _, ok := t.values[code]; ok { return code, true }
    return "", false
}
func (t *GenderType) Validate(value string) bool { _, ok := t.values[value]; return ok }

// JsonType packs/unpacks JSON values, not matchable.
type JsonType struct{ BaseType }
func NewJsonType() *JsonType { return &JsonType{BaseType{name: "json", label: "Nested data", matchable: false}} }
func (t *JsonType) Validate(value string) bool { var v any; return json.Unmarshal([]byte(value), &v) == nil }
func (t *JsonType) Clean(raw string, _ bool, _ string, _ *EntityProxy) (string, bool) {
    s, ok := sanitizeText(raw); if !ok { return "", false }
    var v any
    if json.Unmarshal([]byte(s), &v) == nil { return s, true }
    b, _ := json.Marshal(s)
    return string(b), true
}
func (t *JsonType) NodeID(string) (string, bool) { return "", false }

// Registry holds known property types and helpers.
type Registry struct {
	// Commonly referenced types
    String  *StringType
    Text    *TextType
    HTML    *HTMLType
    Name    *NameType
    Date    *DateType
    Number  *NumberType
    URL     *URLType
    Country *CountryType
    Email   *EmailType
    IP      *IpType
    Phone   *PhoneType
    Address *AddressType
    Language *LanguageType
    Mime    *MimeType
    Checksum *ChecksumType
    Identifier *IdentifierType
    Entity  *EntityType
    Topic   *TopicType
    Gender  *GenderType
    Json    *JsonType

	types     map[string]PropertyType
	matchable map[string]PropertyType
	pivots    map[string]PropertyType
	groups    map[string]PropertyType
}

func NewRegistry() *Registry {
    r := &Registry{
        String:    NewStringType(),
        Text:      NewTextType(),
        HTML:      NewHTMLType(),
        Name:      NewNameType(),
        Date:      NewDateType(),
        Number:    NewNumberType(),
        URL:       NewURLType(),
        Country:   NewCountryType(),
        Email:     NewEmailType(),
        IP:        NewIpType(),
        Phone:     NewPhoneType(),
        Address:   NewAddressType(),
        Language:  NewLanguageType(),
        Mime:      NewMimeType(),
        Checksum:  NewChecksumType(),
        Identifier: NewIdentifierType(),
        Entity:    NewEntityType(),
        Topic:     NewTopicType(),
        Gender:    NewGenderType(),
        Json:      NewJsonType(),
        types:     map[string]PropertyType{},
        matchable: map[string]PropertyType{},
        pivots:    map[string]PropertyType{},
        groups:    map[string]PropertyType{},
    }
    for _, t := range []PropertyType{r.String, r.Text, r.HTML, r.Name, r.Date, r.Number, r.URL, r.Country, r.Email, r.IP, r.Phone, r.Address, r.Language, r.Mime, r.Checksum, r.Identifier, r.Entity, r.Topic, r.Gender, r.Json} {
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
