package ftm

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

// BaseID is the property name used to encode the entity ID as a statement.
const BaseID = "id"

// Statement represents a single assertion about an entity property.
// Fields are modeled after the Python implementation.
type Statement struct {
	ID          string `json:"id,omitempty"`
	EntityID    string `json:"entity_id"`
	CanonicalID string `json:"canonical_id,omitempty"`
	Prop        string `json:"prop"`
	Schema      string `json:"schema"`
	Value       string `json:"value"`
	Dataset     string `json:"dataset"`
	Lang        string `json:"lang,omitempty"`
	Original    string `json:"original_value,omitempty"`
	External    bool   `json:"external"`
	FirstSeen   string `json:"first_seen,omitempty"`
	LastSeen    string `json:"last_seen,omitempty"`
	Origin      string `json:"origin,omitempty"`
}

// MakeKey computes a deterministic ID for a statement.
func (s *Statement) MakeKey() string {
	s.ID = MakeStatementKey(s.Dataset, s.EntityID, s.Prop, s.Value, s.External)
	return s.ID
}

// MakeStatementKey hashes the key properties to produce an ID.
func MakeStatementKey(dataset, entityID, prop, value string, external bool) string {
	if prop == "" || value == "" {
		return ""
	}
	key := fmt.Sprintf("%s.%s.%s.%s", dataset, entityID, prop, value)
	if external {
		key += ".ext"
	}
	h := sha1.Sum([]byte(key))
	return hex.EncodeToString(h[:])
}

// PropTypeName resolves the property type name for a (schema, prop) pair.
// Returns BaseID for the BaseID property.
func PropTypeName(m *Model, schema, prop string) (string, error) {
	if prop == BaseID {
		return BaseID, nil
	}
	sc := m.Get(schema)
	if sc == nil {
		return "", fmt.Errorf("schema not found: %s", schema)
	}
	pr := sc.Get(prop)
	if pr == nil {
		return "", fmt.Errorf("property not found: %s", prop)
	}
	return pr.Type.Name(), nil
}

// StatementsFromEntity emits statements for an entity.
func StatementsFromEntity(e *EntityProxy, dataset string, firstSeen, lastSeen string, external bool, origin string) []Statement {
	if e == nil || e.ID == "" {
		return nil
	}
	st := make([]Statement, 0, 1+len(e.props))
	base := Statement{
		EntityID:    e.ID,
		CanonicalID: e.ID,
		Prop:        BaseID,
		Schema:      e.Schema.Name,
		Value:       e.ID,
		Dataset:     dataset,
		External:    external,
		FirstSeen:   firstSeen,
		LastSeen:    ifEmpty(lastSeen, firstSeen),
		Origin:      origin,
	}
	base.MakeKey()
	st = append(st, base)

	for name, vals := range e.props {
		for _, v := range vals {
			s := Statement{
				EntityID:    e.ID,
				CanonicalID: e.ID,
				Prop:        name,
				Schema:      e.Schema.Name,
				Value:       v,
				Dataset:     dataset,
				External:    external,
				FirstSeen:   firstSeen,
				LastSeen:    ifEmpty(lastSeen, firstSeen),
				Origin:      origin,
			}
			s.MakeKey()
			st = append(st, s)
		}
	}
	return st
}

func ifEmpty(v, alt string) string {
	if v == "" {
		return alt
	}
	return v
}

// GroupKey returns canonical_id if present, otherwise entity_id.
func (s *Statement) GroupKey() string {
	if s.CanonicalID != "" {
		return s.CanonicalID
	}
	return s.EntityID
}

// Clean normalizes trivial fields (e.g., trim, last_seen fallback) without type cleaning.
func (s *Statement) Clean() {
	s.EntityID = strings.TrimSpace(s.EntityID)
	if s.CanonicalID == "" {
		s.CanonicalID = s.EntityID
	}
	if s.LastSeen == "" {
		s.LastSeen = s.FirstSeen
	}
}
