package ftm

// StatementAggregator does streaming aggregation assuming input statements are ordered by GroupKey.
type StatementAggregator struct {
	m   *Model
	cur *EntityProxy
	key string
}

func NewStatementAggregator(m *Model) *StatementAggregator { return &StatementAggregator{m: m} }

// Add consumes one statement. If the group key changes, it returns the completed entity for the previous group.
func (sa *StatementAggregator) Add(s Statement) *EntityProxy {
	gk := s.GroupKey()
	if sa.cur == nil || gk != sa.key {
		// return previous
		var done *EntityProxy
		if sa.cur != nil {
			done = sa.cur
		}
		sc := sa.m.Get(s.Schema)
		if sc == nil {
			return done
		}
		sa.cur = NewEntityProxy(sc, gk)
		sa.key = gk
		if s.Prop != BaseID {
			_ = sa.cur.Add(s.Prop, []string{s.Value}, true, "")
		}
		return done
	}
	if s.Prop != BaseID {
		_ = sa.cur.Add(s.Prop, []string{s.Value}, true, "")
	}
	return nil
}

// Flush returns the current entity, if any.
func (sa *StatementAggregator) Flush() *EntityProxy {
	done := sa.cur
	sa.cur = nil
	sa.key = ""
	return done
}
