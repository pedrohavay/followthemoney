package ftm

import (
	"sort"
)

// AggregateSortedStatements aggregates a slice of statements assumed to be sorted by GroupKey
// (canonical_id or entity_id). It returns a slice of EntityProxy constructed by merging
// statements for each group.
func AggregateSortedStatements(m *Model, st []Statement) []*EntityProxy {
	if len(st) == 0 {
		return nil
	}
	// Ensure sorted by GroupKey for safety
	sort.Slice(st, func(i, j int) bool { return st[i].GroupKey() < st[j].GroupKey() })
	var out []*EntityProxy
	var cur *EntityProxy
	var curKey string
	for i := range st {
		s := st[i]
		key := s.GroupKey()
		if cur == nil || key != curKey {
			if cur != nil {
				out = append(out, cur)
			}
			// Start a new entity using schema from statement
			sc := m.Get(s.Schema)
			if sc == nil {
				continue
			}
			cur = NewEntityProxy(sc, key)
			curKey = key
		}
		if s.Prop == BaseID {
			// We already set ID to group key; ignore base ID
			continue
		}
		// Add property value (cleaned assumed)
		_ = cur.Add(s.Prop, []string{s.Value}, true, "")
	}
	if cur != nil {
		out = append(out, cur)
	}
	return out
}
