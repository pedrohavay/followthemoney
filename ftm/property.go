package ftm

// Property models a schema field, including type and constraints.
// Reverse properties are stubs created implicitly for inbound edges.
type Property struct {
	Schema *Schema
	Name   string
	QName  string

	Label       string
	Description string
	Hidden      bool
	Matchable   bool
	Deprecated  bool
	MaxLength   int

	Type   PropertyType
	Range  *Schema
	Format string

	// Reverse stub information
	Stub    bool
	Reverse *Property
}

type reverseSpec struct {
	Name   string `yaml:"name" json:"name"`
	Label  string `yaml:"label" json:"label"`
	Hidden *bool  `yaml:"hidden" json:"hidden"`
}

type propertySpec struct {
	Label       string       `yaml:"label" json:"label"`
	Description string       `yaml:"description" json:"description"`
	Type        string       `yaml:"type" json:"type"`
	Hidden      *bool        `yaml:"hidden" json:"hidden"`
	Matchable   *bool        `yaml:"matchable" json:"matchable"`
	Deprecated  *bool        `yaml:"deprecated" json:"deprecated"`
	MaxLength   *int         `yaml:"maxLength" json:"maxLength"`
	Range       string       `yaml:"range" json:"range"`
	Format      string       `yaml:"format" json:"format"`
	Reverse     *reverseSpec `yaml:"reverse" json:"reverse"`
}

func newProperty(schema *Schema, name string, spec propertySpec) (*Property, error) {
	p := &Property{
		Schema:      schema,
		Name:        name,
		QName:       schema.Name + ":" + name,
		Label:       spec.Label,
		Description: spec.Description,
		Hidden:      spec.Hidden != nil && *spec.Hidden,
		Deprecated:  spec.Deprecated != nil && *spec.Deprecated,
		MaxLength:   0,
		Format:      spec.Format,
	}
	if spec.MaxLength != nil {
		p.MaxLength = *spec.MaxLength
	}
	tName := spec.Type
	if tName == "" {
		tName = "string"
	}
	p.Type = registry.Get(tName)
	if p.Type == nil {
		// Fallback to string type for unsupported types in this minimal port.
		p.Type = registry.String
	}
	if spec.Matchable != nil {
		p.Matchable = *spec.Matchable
	} else {
		p.Matchable = p.Type.Matchable()
	}
	return p, nil
}
