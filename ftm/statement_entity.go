package ftm

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
)

// StatementEntity stores statements grouped by property for a single canonical entity.
// It acts like an EntityProxy but preserves provenance as statements.
type StatementEntity struct {
    Schema *Schema
    ID     string

	Dataset string // default dataset name for new statements

	// prop -> map[statementID]Statement
	stmts map[string]map[string]Statement

	ExtraReferents map[string]struct{}
	LastChange     string
}

func NewStatementEntity(m *Model, dataset string, schemaName string, id string) (*StatementEntity, error) {
	sc := m.Get(schemaName)
	if sc == nil {
		return nil, fmt.Errorf("schema not found: %s", schemaName)
	}
	return &StatementEntity{Schema: sc, ID: id, Dataset: dataset, stmts: map[string]map[string]Statement{}, ExtraReferents: map[string]struct{}{}}, nil
}

// AddStatement adds a statement to the entity, adapting schema if needed.
func (se *StatementEntity) AddStatement(m *Model, s Statement) error {
	// Adjust schema to common schema if mismatched
	if se.Schema.Name != s.Schema && !se.Schema.IsA(s.Schema) {
		other := m.Get(s.Schema)
		if other == nil {
			return fmt.Errorf("schema not found: %s", s.Schema)
		}
		cs, err := m.CommonSchema(se.Schema, other)
		if err != nil {
			return err
		}
		se.Schema = cs
	}
	if s.Prop == BaseID {
		if s.FirstSeen != "" {
			if se.LastChange == "" || s.FirstSeen > se.LastChange {
				se.LastChange = s.FirstSeen
			}
		}
		return nil
	}
    if se.stmts[s.Prop] == nil {
        se.stmts[s.Prop] = map[string]Statement{}
    }
    if s.ID == "" {
        s.MakeKey()
    }
    if s.PropType == "" {
        if t, err := PropTypeName(m, s.Schema, s.Prop); err == nil {
            s.PropType = t
        }
    }
    // keep canonical id aligned if provided
    if s.CanonicalID == "" && se.ID != "" {
        s.CanonicalID = se.ID
    }
    se.stmts[s.Prop][s.ID] = s
	if s.EntityID != "" && s.EntityID != se.ID {
		se.ExtraReferents[s.EntityID] = struct{}{}
	}
	return nil
}

// Add cleaned value as a statement for the given property.
func (se *StatementEntity) Add(m *Model, propName, value, lang, original, origin string, seen string) error {
    prop := se.Schema.Get(propName)
    if prop == nil {
        return fmt.Errorf("invalid property: %s", propName)
    }
	// Clean via type
	clean, ok := prop.Type.Clean(value, false, prop.Format, nil)
	if !ok || clean == "" {
		return nil
	}
    stmt := Statement{
        EntityID:    se.ID,
        CanonicalID: se.ID,
        Prop:        prop.Name,
        PropType:    prop.Type.Name(),
        Schema:      se.Schema.Name,
        Value:       clean,
        Dataset:     se.Dataset,
        Lang:        lang,
        Original:    original,
		FirstSeen:   seen,
		Origin:      origin,
	}
	stmt.MakeKey()
	return se.AddStatement(m, stmt)
}

// Statements returns all statements, including a synthetic BaseID checksum statement.
func (se *StatementEntity) Statements() []Statement {
    out := make([]Statement, 0)
    ids := make([]string, 0)
    lastSeen := ""
    firstSeen := ""
	// deterministic order
	keys := make([]string, 0, len(se.stmts))
	for k := range se.stmts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
    for _, k := range keys {
        // order by statement id for checksum stability
        arr := make([]Statement, 0, len(se.stmts[k]))
        for _, s := range se.stmts[k] {
            arr = append(arr, s)
        }
        sort.Slice(arr, func(i, j int) bool { return arr[i].ID < arr[j].ID })
        for _, s := range arr {
            out = append(out, s)
            if s.ID != "" {
                ids = append(ids, s.ID)
            }
            if s.LastSeen != "" && s.LastSeen > lastSeen {
                lastSeen = s.LastSeen
            }
            if firstSeen == "" || (s.FirstSeen != "" && s.FirstSeen < firstSeen) {
                firstSeen = s.FirstSeen
            }
        }
    }
    if se.ID != "" {
        out = append(out, Statement{
            ID:          "", // will be computed by MakeKey on write paths if needed
            EntityID:    se.ID,
            CanonicalID: se.ID,
            Prop:        BaseID,
            PropType:    BaseID,
            Schema:      se.Schema.Name,
            Value:       se.ID,
            Dataset:     se.Dataset,
            FirstSeen:   ifEmpty(firstSeen, se.LastChange),
            LastSeen:    lastSeen,
        })
    }
    return out
}

func (se *StatementEntity) Referents() []string {
	out := make([]string, 0, len(se.ExtraReferents))
	for id := range se.ExtraReferents {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
