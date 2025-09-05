package ftm

import (
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// WriteStatementsMsgpack writes statements in MessagePack format as an array stream.
func WriteStatementsMsgpack(w io.Writer, st []Statement) error {
    enc := msgpack.NewEncoder(w)
    // write array header
    if err := enc.EncodeArrayLen(len(st)); err != nil {
        return err
    }
    for i := range st {
        st[i].Clean()
        if st[i].ID == "" {
            st[i].MakeKey()
        }
        if st[i].PropType == "" {
            if t, err := PropTypeName(Default(), st[i].Schema, st[i].Prop); err == nil {
                st[i].PropType = t
            }
        }
        if err := enc.Encode(st[i]); err != nil {
            return err
        }
    }
    return nil
}

// ReadStatementsMsgpack reads statements encoded as an array.
func ReadStatementsMsgpack(r io.Reader, fn func(Statement) error) error {
    dec := msgpack.NewDecoder(r)
    n, err := dec.DecodeArrayLen()
    if err != nil {
        return err
    }
    for i := 0; i < n; i++ {
        var s Statement
        if err := dec.Decode(&s); err != nil {
            return err
        }
        s.Clean()
        if s.ID == "" {
            s.MakeKey()
        }
        if s.PropType == "" {
            if t, err := PropTypeName(Default(), s.Schema, s.Prop); err == nil {
                s.PropType = t
            }
        }
        if err := fn(s); err != nil {
            return err
        }
    }
    return nil
}
