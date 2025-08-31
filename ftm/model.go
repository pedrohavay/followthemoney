package ftm

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Model holds all schema definitions and helpers.
type Model struct {
	Path       string
	Schemata   map[string]*Schema
	Properties map[string]*Property // set of all properties (by qname)
	QNames     map[string]*Property

	// indexes to resolve cross-links during Generate
	extendsIndex map[string][]*Schema
	rangeIndex   map[string]string      // prop.qname -> schema name
	reverseIndex map[string]reverseSpec // prop.qname -> reverseSpec

	once sync.Once
}

func NewModel(path string) (*Model, error) {
	m := &Model{
		Path:         path,
		Schemata:     map[string]*Schema{},
		Properties:   map[string]*Property{},
		QNames:       map[string]*Property{},
		extendsIndex: map[string][]*Schema{},
		rangeIndex:   map[string]string{},
		reverseIndex: map[string]reverseSpec{},
	}
	if err := m.loadAll(); err != nil {
		return nil, err
	}
	if err := m.Generate(); err != nil {
		return nil, err
	}
	return m, nil
}

// Instance returns a singleton model, loading from env FTM_MODEL_PATH or ./followthemoney/schema.
var defaultModel *Model

func Instance() *Model {
	var err error
	if defaultModel == nil {
		path := os.Getenv("FTM_MODEL_PATH")
		if path == "" {
			candidates := []string{
				"schema",
				filepath.Join("goftm", "schema"),
				filepath.Join("..", "schema"),
			}
			if exe, exErr := os.Executable(); exErr == nil {
				base := filepath.Dir(exe)
				candidates = append([]string{filepath.Join(base, "schema")}, candidates...)
			}
			for _, c := range candidates {
				if st, err := os.Stat(c); err == nil && st.IsDir() {
					path = c
					break
				}
			}
			if path == "" {
				path = "schema"
			}
		}
		defaultModel, err = NewModel(path)
		if err != nil {
			// As a fallback, try current directory; otherwise panic to surface configuration error.
			panic(fmt.Errorf("failed to load FtM model from %s: %w", path, err))
		}
	}
	return defaultModel
}

func (m *Model) loadAll() error {
	// Walk all YAML files and load schemata into the model
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".yml") && !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}
		// parse yaml file
		raw, err := os.ReadFile(path)
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
			if _, ok := m.Schemata[name]; ok {
				return fmt.Errorf("duplicate schema name: %s", name)
			}
			m.Schemata[name] = sc
			// capture extends relations (names only; resolved later)
			if len(spec.Extends) > 0 {
				// record on child which parents it extends; resolve later
				// We'll store names first in an aux index: m.extendsIndex will be filled later
				// For now, we just save the names in a temporary property on the schema via model index
				// We'll resolve to schema pointers in Generate
				// We'll store as names on a separate map keyed by child name using temporary schema object
				// Add during Generate by reading spec again: to avoid re-parsing the file we keep them here
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
	if err := filepath.WalkDir(m.Path, walk); err != nil {
		return err
	}

	// Second pass: build extends index from re-reading files (to get parents list) or infer from loaded schemata.
	// Since we already have m.Schemata, we can walk again and fill extendsIndex using schemaSpec.Extends.
	if err := filepath.WalkDir(m.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".yml") && !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fileDefs := map[string]schemaSpec{}
		if err := yaml.Unmarshal(raw, &fileDefs); err != nil {
			return err
		}
		for name, spec := range fileDefs {
			// For each parent name, find schema and append to extendsIndex[child]
			for _, parentName := range spec.Extends {
				if parent := m.Schemata[parentName]; parent != nil {
					m.extendsIndex[name] = append(m.extendsIndex[name], parent)
				} else {
					return fmt.Errorf("invalid extends: %s -> %s", name, parentName)
				}
			}
		}
		return nil
	}); err != nil {
		return err
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

func (m *Model) Get(name string) *Schema { return m.Schemata[name] }

// DefaultModel is the default, package-level instance of the FtM model.
// It loads schemata from the local "schema" folder by default, or from
// the path specified via the FTM_MODEL_PATH environment variable.
//
// You can override it in your application by assigning a custom instance:
//
//	m, _ := ftm.NewModel("/path/to/schema")
//	ftm.DefaultModel = m
var DefaultModel = Instance()
