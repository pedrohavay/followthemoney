package ftm

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io"
	"strconv"
)

// WriteStatementsJSONL writes statements as JSON lines.
func WriteStatementsJSONL(w io.Writer, st []Statement) error {
	enc := json.NewEncoder(w)
	for i := range st {
		st[i].Clean()
		if st[i].ID == "" {
			st[i].MakeKey()
		}
		if err := enc.Encode(&st[i]); err != nil {
			return err
		}
	}
	return nil
}

// ReadStatementsJSONL reads statements from a JSON lines stream.
func ReadStatementsJSONL(r io.Reader, fn func(Statement) error) error {
	dec := json.NewDecoder(bufio.NewReader(r))
	for {
		var s Statement
		if err := dec.Decode(&s); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		s.Clean()
		if s.ID == "" {
			s.MakeKey()
		}
		if err := fn(s); err != nil {
			return err
		}
	}
}

// WriteStatementsCSV a minimal CSV writer (header with common fields).
func WriteStatementsCSV(w io.Writer, st []Statement) error {
	cw := csv.NewWriter(w)
	header := []string{"id", "entity_id", "canonical_id", "prop", "schema", "value", "dataset", "lang", "original_value", "external", "first_seen", "last_seen", "origin"}
	if err := cw.Write(header); err != nil {
		return err
	}
	rec := make([]string, len(header))
	for i := range st {
		s := st[i]
		s.Clean()
		if s.ID == "" {
			s.MakeKey()
		}
		rec[0] = s.ID
		rec[1] = s.EntityID
		rec[2] = s.CanonicalID
		rec[3] = s.Prop
		rec[4] = s.Schema
		rec[5] = s.Value
		rec[6] = s.Dataset
		rec[7] = s.Lang
		rec[8] = s.Original
		if s.External {
			rec[9] = "true"
		} else {
			rec[9] = "false"
		}
		rec[10] = s.FirstSeen
		rec[11] = s.LastSeen
		rec[12] = s.Origin
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// ReadStatementsCSV reads statements from a CSV reader with the same header as WriteStatementsCSV
// and calls fn for each parsed statement.
func ReadStatementsCSV(r io.Reader, fn func(Statement) error) error {
	cr := csv.NewReader(bufio.NewReader(r))
	header, err := cr.Read()
	if err != nil {
		return err
	}
	idx := map[string]int{}
	for i, h := range header {
		idx[h] = i
	}
	get := func(rec []string, key string) string {
		if p, ok := idx[key]; ok && p < len(rec) {
			return rec[p]
		}
		return ""
	}
	for {
		rec, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		s := Statement{
			ID:          get(rec, "id"),
			EntityID:    get(rec, "entity_id"),
			CanonicalID: get(rec, "canonical_id"),
			Prop:        get(rec, "prop"),
			Schema:      get(rec, "schema"),
			Value:       get(rec, "value"),
			Dataset:     get(rec, "dataset"),
			Lang:        get(rec, "lang"),
			Original:    get(rec, "original_value"),
			FirstSeen:   get(rec, "first_seen"),
			LastSeen:    get(rec, "last_seen"),
			Origin:      get(rec, "origin"),
		}
		if p, ok := idx["external"]; ok && p < len(rec) {
			b, _ := strconv.ParseBool(rec[p])
			s.External = b
		}
		s.Clean()
		if s.ID == "" {
			s.MakeKey()
		}
		if err := fn(s); err != nil {
			return err
		}
	}
}
