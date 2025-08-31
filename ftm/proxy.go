package ftm

import (
	"errors"
	"fmt"
	"sort"
)

// EntityProxy wraps an entity instance with its schema and property values.
// It provides validation, normalization, and utility methods.
type EntityProxy struct {
	Schema    *Schema
	ID        string
	KeyPrefix string
	Context   map[string]any // passthrough contextual fields

	props map[string][]string
	size  int // accumulated size of string values
}

func NewEntityProxy(schema *Schema, id string) *EntityProxy {
	return &EntityProxy{Schema: schema, ID: id, Context: map[string]any{}, props: map[string][]string{}}
}

// MakeID creates a hashed ID from the provided parts and key prefix.
func (e *EntityProxy) MakeID(parts ...string) (string, bool) {
	id, ok := makeEntityID(e.KeyPrefix, parts...)
	if ok {
		e.ID = id
	}
	return e.ID, ok
}

func (e *EntityProxy) getProp(name string, quiet bool) (*Property, error) {
	if p := e.Schema.Get(name); p != nil {
		return p, nil
	}
	if quiet {
		return nil, nil
	}
	return nil, errors.New("unknown property: " + name)
}

// Get returns all values for a property by name.
func (e *EntityProxy) Get(name string, quiet bool) []string {
	if _, err := e.getProp(name, quiet); err != nil {
		return nil
	}
	xs := e.props[name]
	out := make([]string, len(xs))
	copy(out, xs)
	return out
}

// First returns the first value for a property.
func (e *EntityProxy) First(name string, quiet bool) string {
	xs := e.Get(name, quiet)
	if len(xs) > 0 {
		return xs[0]
	}
	return ""
}

// Has tests if a property has at least one value.
func (e *EntityProxy) Has(name string, quiet bool) bool { _, ok := e.props[name]; return ok }

// Add adds (and normalizes) values for a property.
func (e *EntityProxy) Add(name string, values []string, fuzzy bool, format string) error {
	p, err := e.getProp(name, false)
	if err != nil || p == nil {
		return err
	}
	if p.Stub {
		return errors.New("stub property cannot be written")
	}
	// iterate and clean
	if e.props[name] == nil {
		e.props[name] = []string{}
	}
	set := map[string]struct{}{}
	for _, v := range e.props[name] {
		set[v] = struct{}{}
	}
	for _, raw := range values {
		clean, ok := p.Type.Clean(raw, fuzzy, p.Format, e)
		if !ok || clean == "" {
			continue
		}
		// aggregate size cap
		if max := p.Type.TotalSize(); max > 0 {
			if e.size+len(clean) > max {
				continue
			}
		}
		if _, seen := set[clean]; !seen {
			e.props[name] = append(e.props[name], clean)
			set[clean] = struct{}{}
			e.size += len(clean)
		}
	}
	return nil
}

// UnsafeAdd is a helper for adding a single already-sanitized value.
func (e *EntityProxy) UnsafeAdd(p *Property, value string, fuzzy bool) (string, bool) {
	clean, ok := p.Type.Clean(value, fuzzy, p.Format, e)
	if !ok || clean == "" {
		return "", false
	}
	if p.Stub {
		return "", false
	}
	if e.props[p.Name] == nil {
		e.props[p.Name] = []string{}
	}
	for _, v := range e.props[p.Name] {
		if v == clean {
			return clean, true
		}
	}
	if max := p.Type.TotalSize(); max > 0 && e.size+len(clean) > max {
		return "", false
	}
	e.props[p.Name] = append(e.props[p.Name], clean)
	e.size += len(clean)
	return clean, true
}

// Set replaces all existing values with the provided ones.
func (e *EntityProxy) Set(name string, values []string, fuzzy bool, format string) error {
	delete(e.props, name)
	return e.Add(name, values, fuzzy, format)
}

// Pop removes all values for a property and returns them.
func (e *EntityProxy) Pop(name string, quiet bool) []string {
	if _, err := e.getProp(name, quiet); err != nil {
		return nil
	}
	xs := e.props[name]
	delete(e.props, name)
	return xs
}

// Remove removes a single value from the property.
func (e *EntityProxy) Remove(name, value string, quiet bool) {
	if _, err := e.getProp(name, quiet); err != nil {
		return
	}
	xs := e.props[name]
	out := xs[:0]
	for _, v := range xs {
		if v != value {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		delete(e.props, name)
	} else {
		e.props[name] = out
	}
}

// IterProps returns properties for which a value is set.
func (e *EntityProxy) IterProps() []*Property {
	props := make([]*Property, 0, len(e.props))
	for name := range e.props {
		if p := e.Schema.Get(name); p != nil {
			props = append(props, p)
		}
	}
	sort.Slice(props, func(i, j int) bool { return props[i].Name < props[j].Name })
	return props
}

// IterValues yields (Property, value) pairs for all values.
func (e *EntityProxy) IterValues() [][2]interface{} {
	pairs := make([][2]interface{}, 0)
	for name, vals := range e.props {
		p := e.Schema.Get(name)
		if p == nil {
			continue
		}
		for _, v := range vals {
			pairs = append(pairs, [2]interface{}{p, v})
		}
	}
	return pairs
}

// EdgePairs returns value pairs for edge source/target if schema represents an edge.
func (e *EntityProxy) EdgePairs() [][2]string {
	if !e.Schema.Edge {
		return nil
	}
	src := e.Get(e.Schema.EdgeSource, true)
	dst := e.Get(e.Schema.EdgeTarget, true)
	out := make([][2]string, 0, len(src)*len(dst))
	for _, s := range src {
		for _, t := range dst {
			out = append(out, [2]string{s, t})
		}
	}
	return out
}

// GetTypeValues returns all values with a given property type name.
func (e *EntityProxy) GetTypeValues(pt PropertyType, matchable bool) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for name, vals := range e.props {
		p := e.Schema.Get(name)
		if p == nil {
			continue
		}
		if matchable && !p.Matchable {
			continue
		}
		if p.Type.Name() == pt.Name() {
			for _, v := range vals {
				if _, ok := seen[v]; ok {
					continue
				}
				seen[v] = struct{}{}
				out = append(out, v)
			}
		}
	}
	return out
}

// Caption picks a human-friendly caption, using schema caption properties.
func (e *EntityProxy) Caption() string {
	// Prefer name-type with multiple values -> heuristic pick (shortest)
	for _, pname := range e.Schema.Caption {
		p := e.Schema.Get(pname)
		if p == nil {
			continue
		}
		values := e.Get(pname, true)
		if p.Type.Name() == registry.Name.Name() && len(values) > 1 {
			return shortest(values...)
		}
		if len(values) > 0 {
			return values[0]
		}
	}
	return e.Schema.Label
}

// Countries returns country-type values set on the entity.
func (e *EntityProxy) Countries() []string { return e.GetTypeValues(registry.Country, false) }

// ToDict serializes the entity to a plain map.
func (e *EntityProxy) ToDict() map[string]any {
	props := map[string][]string{}
	for k, v := range e.props {
		vv := make([]string, len(v))
		copy(vv, v)
		props[k] = vv
	}
	data := map[string]any{
		"id":         e.ID,
		"schema":     e.Schema.Name,
		"properties": props,
	}
	for k, v := range e.Context {
		data[k] = v
	}
	return data
}

// Clone deep-copies the entity proxy.
func (e *EntityProxy) Clone() *EntityProxy {
	cp := NewEntityProxy(e.Schema, e.ID)
	cp.KeyPrefix = e.KeyPrefix
	cp.Context = map[string]any{}
	for k, v := range e.Context {
		cp.Context[k] = v
	}
	for k, vals := range e.props {
		vv := make([]string, len(vals))
		copy(vv, vals)
		cp.props[k] = vv
	}
	cp.size = e.size
	return cp
}

// Merge another entity into this one using most specific common schema.
func (e *EntityProxy) Merge(other *EntityProxy) (*EntityProxy, error) {
	e.ID = firstNonEmpty(e.ID, other.ID)
	schema, err := e.Schema.Model.CommonSchema(e.Schema, other.Schema)
	if err != nil {
		return nil, fmtError("cannot merge entities: %w", err)
	}
	e.Schema = schema
	// merge context (concat unique)
	for k, v := range other.Context {
		if _, ok := e.Context[k]; !ok {
			e.Context[k] = v
		}
	}
	for name, values := range other.props {
		_ = e.Add(name, values, true, "")
	}
	return e, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
func fmtError(format string, a ...any) error { return fmt.Errorf(format, a...) }
