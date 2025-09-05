package ftm

import (
	"fmt"
	"slices"
)

// EdgeSpec defines how a schema is represented as a graph edge.
type EdgeSpec struct {
	Source   string   `yaml:"source" json:"source"`
	Target   string   `yaml:"target" json:"target"`
	Caption  []string `yaml:"caption" json:"caption"`
	Label    string   `yaml:"label" json:"label"`
	Directed *bool    `yaml:"directed" json:"directed"`
}

type TemporalExtentSpec struct {
	Start []string `yaml:"start" json:"start"`
	End   []string `yaml:"end" json:"end"`
}

// Schema models an entity class with properties and inheritance.
type Schema struct {
	Model *Model
	Name  string

	Label       string
	Plural      string
	Description string

	Abstract   bool
	Hidden     bool
	Generated  bool
	Matchable  bool
	Deprecated bool

	Featured []string
	Required []string
	Caption  []string

	EdgeSpec       EdgeSpec
	TemporalExtent TemporalExtentSpec

	// Resolved graph semantics
	Edge         bool
	EdgeDirected bool
	EdgeSource   string
	EdgeTarget   string
	EdgeCaption  []string
	edgeLabel    string

	Extends     []*Schema
	Schemata    map[string]*Schema  // includes self & ancestors
	Names       map[string]struct{} // names of Schemata
	Descendants map[string]*Schema

	Properties map[string]*Property

	temporalStart []string
	temporalEnd   []string

	generated bool
}

type schemaSpec struct {
	Label       string                  `yaml:"label" json:"label"`
	Plural      string                  `yaml:"plural" json:"plural"`
	Schemata    []string                `yaml:"schemata" json:"schemata"`
	Extends     []string                `yaml:"extends" json:"extends"`
	Properties  map[string]propertySpec `yaml:"properties" json:"properties"`
	Featured    []string                `yaml:"featured" json:"featured"`
	Required    []string                `yaml:"required" json:"required"`
	Caption     []string                `yaml:"caption" json:"caption"`
	Edge        EdgeSpec                `yaml:"edge" json:"edge"`
	Temporal    TemporalExtentSpec      `yaml:"temporalExtent" json:"temporalExtent"`
	Description string                  `yaml:"description" json:"description"`
	Abstract    *bool                   `yaml:"abstract" json:"abstract"`
	Hidden      *bool                   `yaml:"hidden" json:"hidden"`
	Generated   *bool                   `yaml:"generated" json:"generated"`
	Matchable   *bool                   `yaml:"matchable" json:"matchable"`
	Deprecated  *bool                   `yaml:"deprecated" json:"deprecated"`
}

func newSchema(m *Model, name string, spec schemaSpec) (*Schema, error) {
	s := &Schema{
		Model:          m,
		Name:           name,
		Label:          spec.Label,
		Plural:         spec.Plural,
		Description:    spec.Description,
		Featured:       append([]string{}, spec.Featured...),
		Required:       append([]string{}, spec.Required...),
		Caption:        append([]string{}, spec.Caption...),
		EdgeSpec:       spec.Edge,
		TemporalExtent: spec.Temporal,
		Extends:        []*Schema{},
		Schemata:       map[string]*Schema{},
		Names:          map[string]struct{}{},
		Descendants:    map[string]*Schema{},
		Properties:     map[string]*Property{},
	}
	if s.Label == "" {
		s.Label = name
	}
	if s.Plural == "" {
		s.Plural = s.Label
	}
	if spec.Abstract != nil {
		s.Abstract = *spec.Abstract
	}
	if spec.Hidden != nil {
		s.Hidden = *spec.Hidden
	}
	if spec.Generated != nil {
		s.Generated = *spec.Generated
	}
    if spec.Matchable != nil {
        s.Matchable = *spec.Matchable
    } else {
        // Default to false when not specified to align with official model semantics.
        s.Matchable = false
    }
	if spec.Deprecated != nil {
		s.Deprecated = *spec.Deprecated
	}

	// Edge basics
	s.EdgeSource = s.EdgeSpec.Source
	s.EdgeTarget = s.EdgeSpec.Target
	s.Edge = s.EdgeSource != "" && s.EdgeTarget != ""
	s.EdgeCaption = append([]string{}, s.EdgeSpec.Caption...)
	s.edgeLabel = s.EdgeSpec.Label
	s.EdgeDirected = true
	if s.EdgeSpec.Directed != nil {
		s.EdgeDirected = *s.EdgeSpec.Directed
	}

	s.temporalStart = append([]string{}, s.TemporalExtent.Start...)
	s.temporalEnd = append([]string{}, s.TemporalExtent.End...)

	// Own properties for now; ranges/reverses resolved in Generate
	for pn, ps := range spec.Properties {
		p, err := newProperty(s, pn, ps)
		if err != nil {
			return nil, err
		}
		s.Properties[pn] = p
	}
	// initialize own ancestry with self
	s.Schemata[s.Name] = s
	s.Names[s.Name] = struct{}{}
	return s, nil
}

// Generate finalizes the schema: resolve inheritance, properties, reverse links, etc.
func (s *Schema) Generate() error {
	if s.generated {
		return nil
	}
	// Resolve extends
	// Note: Model calls Generate after loading all schemata; we resolve here lazily.
	// Inherit properties and ancestry.
	for _, parent := range s.Model.extendsIndex[s.Name] {
		// Ensure parent is generated first
		_ = parent.Generate()
		// Parent already exists by model load stage
		if _, ok := s.Schemata[parent.Name]; !ok {
			s.Extends = append(s.Extends, parent)
			for name, prop := range parent.Properties {
				if _, ok := s.Properties[name]; !ok {
					s.Properties[name] = prop
				}
			}
			// Inherit ancestry
			for n, sc := range parent.Schemata {
				s.Schemata[n] = sc
				s.Names[n] = struct{}{}
				sc.Descendants[s.Name] = s
			}
		}
	}

	// Resolve ranges and reverse stubs for entity properties
	for _, prop := range s.Properties {
		if prop.Type.Name() == registry.Entity.Name() {
			if prop.Range == nil {
				if rngName := s.Model.rangeIndex[prop.QName]; rngName != "" {
					if rng := s.Model.Schemata[rngName]; rng != nil {
						prop.Range = rng
					}
				}
			}
			if prop.Reverse == nil {
				if rs, ok := s.Model.reverseIndex[prop.QName]; ok && prop.Range != nil {
					// Create or get reverse property on the range schema
					targetSchema := prop.Range
					// If not exists on target, create stub
					rev := targetSchema.Properties[rs.Name]
					if rev == nil {
						hidden := prop.Hidden
						if rs.Hidden != nil {
							hidden = *rs.Hidden
						}
						rev = &Property{
							Schema: targetSchema,
							Name:   rs.Name,
							QName:  targetSchema.Name + ":" + rs.Name,
							Label:  rs.Label,
							Hidden: hidden,
							Type:   registry.Entity,
							Range:  s,
							Stub:   true,
						}
						targetSchema.Properties[rs.Name] = rev
					}
					prop.Reverse = rev
				}
			}
		}
	}

	s.generated = true
	return nil
}

func (s *Schema) Get(name string) *Property { return s.Properties[name] }

// IsA checks if the schema or any parent matches the candidate name.
func (s *Schema) IsA(candidate string) bool {
	_, ok := s.Names[candidate]
	return ok
}

// SortedProperties returns properties sorted with caption/featured priority then by label.
func (s *Schema) SortedProperties() []*Property {
	props := make([]*Property, 0, len(s.Properties))
	for _, p := range s.Properties {
		props = append(props, p)
	}
	// Keep deterministic order: captions first, then featured, then name
	slices.SortFunc(props, func(a, b *Property) int {
		// caption priority
		ac := indexOf(s.Caption, a.Name)
		bc := indexOf(s.Caption, b.Name)
		if ac != bc {
			return compareIndex(ac, bc)
		}
		// featured priority
		af := indexOf(s.Featured, a.Name)
		bf := indexOf(s.Featured, b.Name)
		if af != bf {
			return compareIndex(af, bf)
		}
		if a.Label < b.Label {
			return -1
		}
		if a.Label > b.Label {
			return 1
		}
		return 0
	})
	return props
}

func indexOf(list []string, val string) int {
	for i, v := range list {
		if v == val {
			return i
		}
	}
	return 1 << 30
}
func compareIndex(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// Temporal properties resolved lists
func (s *Schema) TemporalStartProps() []*Property {
	names := s.temporalStart
	if len(names) == 0 {
		for _, parent := range s.Extends {
			ps := parent.TemporalStartProps()
			if len(ps) > 0 {
				return ps
			}
		}
	}
	props := make([]*Property, 0, len(names))
	for _, n := range names {
		if p := s.Get(n); p != nil {
			props = append(props, p)
		}
	}
	return props
}

func (s *Schema) TemporalEndProps() []*Property {
	names := s.temporalEnd
	if len(names) == 0 {
		for _, parent := range s.Extends {
			ps := parent.TemporalEndProps()
			if len(ps) > 0 {
				return ps
			}
		}
	}
	props := make([]*Property, 0, len(names))
	for _, n := range names {
		if p := s.Get(n); p != nil {
			props = append(props, p)
		}
	}
	return props
}

// Validate checks property presence and basic type validation.
func (s *Schema) Validate(data map[string][]string) error {
	// Required fields present?
	for _, req := range s.Required {
		if len(data[req]) == 0 {
			return fmt.Errorf("required property missing: %s", req)
		}
	}
	// Type-level validation
	for name, values := range data {
		p := s.Properties[name]
		if p == nil {
			continue
		}
		for _, v := range values {
			if !p.Type.Validate(v) {
				return fmt.Errorf("invalid value for %s", name)
			}
		}
	}
	return nil
}
