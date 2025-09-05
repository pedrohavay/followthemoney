# Follow The Money (Go)

[![CI](https://github.com/pedrohavay/followthemoney/actions/workflows/ci.yml/badge.svg)](https://github.com/pedrohavay/followthemoney/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/pedrohavay/followthemoney/ftm.svg)](https://pkg.go.dev/github.com/pedrohavay/followthemoney/ftm)
[![Release](https://img.shields.io/github/v/release/pedrohavay/followthemoney?display_name=tag&sort=semver)](https://github.com/pedrohavay/followthemoney/releases)

A Go port of FollowTheMoney (FtM) — a pragmatic data model for people, companies, assets, relationships and documents
used in investigative work and financial crime analysis.

- Inspiration project (Python): https://github.com/opensanctions/followthemoney
- Documentation reference: https://followthemoney.tech

This library focuses on the FtM core for high‑throughput data pipelines:

- YAML‑based schemata and properties
- Strongly‑typed property registry with real cleaning/validation
- Entity proxies (create, clean, merge, serialize)
- Namespace signing for entity IDs (HMAC)
- Property graph projection (nodes/edges)
- Statements (atomic property assertions) with JSONL/CSV/MessagePack I/O
- In‑memory and streaming aggregation of statements into entities

## Status

- Ready for pipelines that ingest multiple sources, normalize values, emit statements and aggregate into entities (index
  where you prefer).
- Mapping (CSV/SQL → entities) and Dataset metadata are not included yet and may land later.

## Installation

Install the module and import the package in your Go code. Go 1.22 or newer is recommended.

```bash
go get github.com/pedrohavay/followthemoney/ftm
```

Then, import:

```go
import "github.com/pedrohavay/followthemoney/ftm"
```

## Quick start

The example below loads the default model, constructs a `Person` entity, cleans a few raw values
using the type system, and prints a JSON representation suitable for storage or transport.

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/pedrohavay/followthemoney/ftm"
)

func main() {
	// Default model loads from ./schema or FTM_MODEL_PATH if set
	m := ftm.Default()

	// Create a Person
	person := m.Get("Person")
	e := ftm.NewEntityProxy(person, "p1")

	// Add raw values (cleaning and de-dup applied)
	_ = e.Add("name", []string{" John  Smith "}, false, "")
	_ = e.Add("nationality", []string{"DE"}, false, "")

	// Serialize
	data := e.ToDict()
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(b))
}

```

## Types, cleaning and validation

FtM’s type registry encapsulates validation and normalization for common value kinds. The examples below show IDNA
processing for emails and format‑specific handling for identifiers like IBAN and BIC/LEI.

```go
r := ftm.NewRegistry()
email, ok := r.Email.Clean("John <j.smith@bücher.de>", false, "", nil) // j.smith@xn--bcher-kva.de
iban, ok2 := r.Identifier.Clean("DE44 5001 0517 5407 3249 31", false, "iban", nil)
```

## Namespace signing

HMAC‑sign entity IDs to create dataset‑scoped identifiers and avoid collisions across sources. Applying a namespace
also rewrites entity‑typed properties to their signed equivalents.

```go
ns := ftm.NewNamespace("dataset-key")

p := ftm.NewEntityProxy(ftm.Default().Get("Passport"), "doc1")
_ = p.Add("holder", []string{"p1"}, true, "")

signed := ns.Apply(p, false)
// signed.ID and entity-typed values are signed
```

## Property graph

Project entities and selected property values into a property graph. Matchable properties (e.g., names, URLs,
countries) become value nodes connected to entity nodes for visual exploration and graph analytics.

```go
package main

import (
	"fmt"

	"github.com/pedrohavay/followthemoney/ftm"
)

func main() {
	m := ftm.Default()
	e := ftm.NewEntityProxy(m.Get("Person"), "p1")
	_ = e.Add("name", []string{"Ana"}, false, "")

	g := ftm.NewGraph(nil)
	g.Add(e)

	for _, n := range g.Nodes() {
		label := n.Value
		if n.Proxy != nil {
			label = n.Proxy.Caption()
		} else {
			// Use type-specific caption (e.g., nice formatting for URLs, names)
			label = n.Type.Caption(n.Value, "")
		}
		fmt.Println("node", n.Type.Name(), label)
	}
	for _, ed := range g.Edges() {
		fmt.Println("edge", ed.SourceID, "->", ed.TargetID, ed.TypeName())
	}
}

```

## Statements & I/O

Statements encode each (entity, property, value) as a separate record. This enables streaming ingest, provenance
tracking, and idempotent processing. The example demonstrates JSONL, CSV, and MessagePack round‑trips.

```go
// Build statements from an entity
st := ftm.StatementsFromEntity(e, "my_dataset", "2025-01-01", "", false, "ingestor-A")

// JSONL
_ = ftm.WriteStatementsJSONL(os.Stdout, st)
_ = ftm.ReadStatementsJSONL(os.Stdin, func (s ftm.Statement) error { /* handle */ return nil })

// CSV (header-compatible reader/writer)
var buf bytes.Buffer
_ = ftm.WriteStatementsCSV(&buf, st)
_ = ftm.ReadStatementsCSV(&buf, func(s ftm.Statement) error { /* handle */ return nil })

// MessagePack (optional)
_ = ftm.WriteStatementsMsgpack(&buf, st)
_ = ftm.ReadStatementsMsgpack(&buf, func (s ftm.Statement) error { return nil })
```

Notes:
- Statements include `prop_type` (e.g., `name`, `country`, `id`). Readers compute it when absent for backward compatibility.
- The BaseID statement (`prop = "id"`) carries the entity ID in `value` across all producers, including `StatementEntity`.

## Aggregation

Reconstruct entities from statements by aggregating on `canonical_id` (or `entity_id` when canonical is absent).
Batch aggregation expects statements sorted by group key; streaming aggregation emits an entity whenever the group
key changes.

```go
// Sort by GroupKey (canonical_id or entity_id)
slices.SortFunc(st, func (a, b ftm.Statement) int {
    if a.GroupKey() < b.GroupKey() { return -1 }
    if a.GroupKey() > b.GroupKey() { return 1 }
    return 0
})
entities := ftm.AggregateSortedStatements(ftm.Default(), st)
```

Streaming aggregation:

```go
agg := ftm.NewStatementAggregator(ftm.Default())

var entities []*ftm.EntityProxy

for _, s := range st {
    if ent := agg.Add(s); ent != nil {
    entities = append(entities, ent)
    }
}

if ent := agg.Flush(); ent != nil {
    entities = append(entities, ent)
}
```

## Statement Entity

Build an entity by accumulating statements and keep provenance:

```go
se, _ := ftm.NewStatementEntity(ftm.Default(), "ds", "Person", "p1")
_ = se.Add(ftm.Default(), "name", "Ana", "", "", "source-A", "2025-01-01")
_ = se.Add(ftm.Default(), "nationality", "br", "", "", "source-B", "2025-02-10")
statements := se.Statements() // includes BaseID checksum
```

## Roadmap

- Dataset metadata (catalog/coverage/resources), Mapping (CSV/SQL → entities).
- Additional comparators and exporters based on usage.

## Contributing

Issues and PRs welcome. Please discuss larger changes in an issue first.

## Credits

Port by [@pedrohavay](http://x.com/pedrohavay), inspirited on the FollowTheMoney Python
project (https://github.com/opensanctions/followthemoney) and the documentation at
https://followthemoney.tech.
