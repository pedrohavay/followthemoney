package ftm

// Registry holds known property types and helpers.
type Registry struct {
	// Commonly referenced types
	String     *StringType
	Text       *TextType
	HTML       *HTMLType
	Name       *NameType
	Date       *DateType
	Number     *NumberType
	URL        *URLType
	Country    *CountryType
	Email      *EmailType
	IP         *IpType
	Phone      *PhoneType
	Address    *AddressType
	Language   *LanguageType
	Mime       *MimeType
	Checksum   *ChecksumType
	Identifier *IdentifierType
	Entity     *EntityType
	Topic      *TopicType
	Gender     *GenderType
	Json       *JsonType

	types     map[string]PropertyType
	matchable map[string]PropertyType
	pivots    map[string]PropertyType
	groups    map[string]PropertyType
}

func NewRegistry() *Registry {
	r := &Registry{
		String:     NewStringType(),
		Text:       NewTextType(),
		HTML:       NewHTMLType(),
		Name:       NewNameType(),
		Date:       NewDateType(),
		Number:     NewNumberType(),
		URL:        NewURLType(),
		Country:    NewCountryType(),
		Email:      NewEmailType(),
		IP:         NewIpType(),
		Phone:      NewPhoneType(),
		Address:    NewAddressType(),
		Language:   NewLanguageType(),
		Mime:       NewMimeType(),
		Checksum:   NewChecksumType(),
		Identifier: NewIdentifierType(),
		Entity:     NewEntityType(),
		Topic:      NewTopicType(),
		Gender:     NewGenderType(),
		Json:       NewJsonType(),
		types:      map[string]PropertyType{},
		matchable:  map[string]PropertyType{},
		pivots:     map[string]PropertyType{},
		groups:     map[string]PropertyType{},
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
