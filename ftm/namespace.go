package ftm

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

// Namespace partitions entity IDs by appending an HMAC signature suffix.
// Format: <plainID>.<sha1-hmac>
type Namespace struct {
	key []byte
}

func NewNamespace(name string) *Namespace {
	if name == "" {
		return &Namespace{key: nil}
	}
	return &Namespace{key: []byte(name)}
}

// Parse splits an entity id into plain and signature parts.
func (ns *Namespace) Parse(entityID string) (plain, sig string) {
	if entityID == "" {
		return "", ""
	}
	parts := strings.Split(entityID, ".")
	if len(parts) < 2 {
		return entityID, ""
	}
	// signature is tail
	return strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
}

func (ns *Namespace) signature(plain string) string {
	if len(ns.key) == 0 || plain == "" {
		return ""
	}
	mac := hmac.New(sha1.New, ns.key)
	mac.Write([]byte(plain))
	return hex.EncodeToString(mac.Sum(nil))
}

// Sign applies the namespace signature to a plain id, or returns the input if key is empty.
func (ns *Namespace) Sign(entityID string) string {
	plain, _ := ns.Parse(entityID)
	if len(ns.key) == 0 {
		return plain
	}
	if plain == "" {
		return ""
	}
	sig := ns.signature(plain)
	if sig == "" {
		return plain
	}
	return plain + "." + sig
}

// Verify checks if an ID carries a valid signature for this namespace.
func (ns *Namespace) Verify(entityID string) bool {
	plain, sig := ns.Parse(entityID)
	if plain == "" || sig == "" {
		return false
	}
	return hmac.Equal([]byte(sig), []byte(ns.signature(plain)))
}

// Apply rewrites an entity proxy to sign the entity id and any referenced entity properties.
func (ns *Namespace) Apply(e *EntityProxy, shallow bool) *EntityProxy {
	cp := e.Clone()
	if cp.ID != "" {
		cp.ID = ns.Sign(cp.ID)
	}
	if shallow {
		return cp
	}
	for name, vals := range cp.props {
		p := cp.Schema.Get(name)
		if p == nil || p.Type.Name() != registry.Entity.Name() {
			continue
		}
		newVals := make([]string, 0, len(vals))
		for _, v := range vals {
			newVals = append(newVals, ns.Sign(v))
		}
		cp.props[name] = newVals
	}
	return cp
}
