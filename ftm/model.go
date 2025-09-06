package ftm

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"

	ftmschema "github.com/pedrohavay/followthemoney/schema"
	"gopkg.in/yaml.v3"
)

// Model holds all schema definitions and helpers.
type Model struct {
	Path       string
	fsys       fs.FS
	Schemata   map[string]*Schema
	Properties map[string]*Property // set of all properties (by qname)
	QNames     map[string]*Property

	// indexes to resolve cross-links during Generate
	extendsIndex map[string][]*Schema
	rangeIndex   map[string]string      // prop.qname -> schema name
	reverseIndex map[string]reverseSpec // prop.qname -> reverseSpec
	extendsNames map[string][]string    // temporary: child -> parent names

	once sync.Once
}

// NewModel loads the model from filesystem path.
func NewModel(path string) (*Model, error) {
	m := &Model{
		Path:         ".",
		fsys:         os.DirFS(path),
		Schemata:     map[string]*Schema{},
		Properties:   map[string]*Property{},
		QNames:       map[string]*Property{},
		extendsIndex: map[string][]*Schema{},
		rangeIndex:   map[string]string{},
		reverseIndex: map[string]reverseSpec{},
		extendsNames: map[string][]string{},
	}

	// Load all schemata from YAML files in the path
	if err := m.loadAll(); err != nil {
		return nil, err
	}

	// Resolve cross-references and inheritance
	if err := m.Generate(); err != nil {
		return nil, err
	}

	return m, nil
}

// NewModelFS loads the model from a generic filesystem, rooted at `root`.
func NewModelFS(fsys fs.FS, root string) (*Model, error) {
	m := &Model{
		Path:         root,
		fsys:         fsys,
		Schemata:     map[string]*Schema{},
		Properties:   map[string]*Property{},
		QNames:       map[string]*Property{},
		extendsIndex: map[string][]*Schema{},
		rangeIndex:   map[string]string{},
		reverseIndex: map[string]reverseSpec{},
		extendsNames: map[string][]string{},
	}
	if err := m.loadAll(); err != nil {
		return nil, err
	}
	if err := m.Generate(); err != nil {
		return nil, err
	}
	return m, nil
}

// Instance returns a singleton model, loading from env FTM_MODEL_PATH or default schemas.
var defaultModel *Model

// Default returns the singleton model instance.
func Default() *Model {
	var err error

	if defaultModel == nil {
		path := os.Getenv("FTM_MODEL_PATH")
		if path != "" {
			defaultModel, err = NewModel(path)
			if err == nil {
				return defaultModel
			}
		}

		// Try embedded schema files
		defaultModel, err = NewModelFS(ftmschema.Files, ".")
		if err != nil {
			// Fallback for development: local folder named "schema"
			defaultModel, err = NewModel("schema")
			if err != nil {
				panic(fmt.Errorf("failed to load FtM model: %w", err))
			}
		}
	}

	return defaultModel
}

// loadAll walks the filesystem and loads all YAML schema files.
func (m *Model) loadAll() error {
	// Walk all YAML files and load schemata into the model
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Skip non-yaml files
		if !strings.HasSuffix(d.Name(), ".yml") && !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}

		// Parse yaml file
		raw, err := fs.ReadFile(m.fsys, path)
		if err != nil {
			return err
		}

		// Each file is a map[name]schemaSpec
		fileDefs := map[string]schemaSpec{}

		if err := yaml.Unmarshal(raw, &fileDefs); err != nil {
			return err
		}

		for name, spec := range fileDefs {
			sc, err := newSchema(m, name, spec)
			if err != nil {
				return err
			}

			// Register schema
			if _, ok := m.Schemata[name]; ok {
				return fmt.Errorf("duplicate schema name: %s", name)
			}
			m.Schemata[name] = sc

			// Capture extends relations (names only; resolved later)
			if len(spec.Extends) > 0 {
				m.extendsNames[name] = append(m.extendsNames[name], spec.Extends...)
			}

			// Prepare per-property range and reverse indexes
			for pn, ps := range spec.Properties {
				qname := name + ":" + pn

				if ps.Range != "" {
					m.rangeIndex[qname] = ps.Range
				}

				if ps.Reverse != nil {
					m.reverseIndex[qname] = *ps.Reverse
				}
			}
		}

		return nil
	}
	if err := fs.WalkDir(m.fsys, m.Path, walk); err != nil {
		return err
	}

	// Resolve extends names into schema pointers now that all schemata are loaded
	for child, parents := range m.extendsNames {
		for _, parentName := range parents {
			parent := m.Schemata[parentName]
			if parent == nil {
				return fmt.Errorf("invalid extends: %s -> %s", child, parentName)
			}

			// Register child -> parent link
			m.extendsIndex[child] = append(m.extendsIndex[child], parent)
		}
	}

	return nil
}

// Generate resolves cross-references and inheritance.
func (m *Model) Generate() error {
	// Resolve schemata inheritance and property reverses/ranges
	for _, s := range m.Schemata {
		if err := s.Generate(); err != nil {
			return err
		}
	}

	// Build QName index and ensure children inherit properties defined on ancestors (already done in Generate)
	for _, s := range m.Schemata {
		for _, p := range s.Properties {
			m.QNames[p.QName] = p
			m.Properties[p.QName] = p
		}
	}

	return nil
}

// CommonSchema selects the most specific of two schemata if comparable.
func (m *Model) CommonSchema(left, right *Schema) (*Schema, error) {
	if left == nil || right == nil {
		return nil, errors.New("invalid schema")
	}

	if left.IsA(right.Name) {
		return left, nil
	}

	if right.IsA(left.Name) {
		return right, nil
	}

	return nil, fmt.Errorf("no common schema: %s and %s", left.Name, right.Name)
}

// Get returns the schema by name, or nil if not found.
func (m *Model) Get(name string) *Schema { return m.Schemata[name] }
