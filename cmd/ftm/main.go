package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pedrohavay/followthemoney/ftm"
)

// Minimal CLI mirroring core Python commands: dump-model, validate, pretty, sign.
// Usage:
//   ftm dump-model
//   ftm validate < infile.jsonl > outfile.jsonl
//   ftm pretty < infile.jsonl
//   ftm sign -key <secret> < infile.jsonl > outfile.jsonl

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	switch cmd {
	case "dump-model":
		dumpModel()
	case "validate":
		validate()
	case "pretty":
		pretty()
	case "sign":
		sign()
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "ftm commands: dump-model | validate | pretty | sign\n")
}

func dumpModel() {
	_ = ftm.Default() // ensure model loads
	// Compact metadata: schemata names list and property qnames
	out := map[string]any{"schemata": map[string]any{}, "types": []string{"string", "text", "name", "date", "number", "url", "country", "entity"}}
	for name, sc := range ftm.Default().Schemata {
		props := map[string]any{}
		for n, p := range sc.Properties {
			props[n] = map[string]any{"name": p.Name, "qname": p.QName, "type": p.Type.Name(), "label": p.Label}
		}
		out["schemata"].(map[string]any)[name] = map[string]any{
			"label":      sc.Label,
			"plural":     sc.Plural,
			"extends":    schemaNames(sc.Extends),
			"properties": props,
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

func schemaNames(xs []*ftm.Schema) []string {
	out := make([]string, 0, len(xs))
	for _, s := range xs {
		out = append(out, s.Name)
	}
	return out
}

type entityJSON struct {
	ID         string              `json:"id"`
	Schema     string              `json:"schema"`
	Properties map[string][]string `json:"properties"`
}

func validate() {
	m := ftm.Default()
	br := bufio.NewReader(os.Stdin)
	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()
	dec := json.NewDecoder(br)
	enc := json.NewEncoder(bw)
	for {
		var e entityJSON
		if err := dec.Decode(&e); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "error decoding JSON: %v\n", err)
			os.Exit(1)
		}
		sc := m.Get(e.Schema)
		if sc == nil {
			fmt.Fprintf(os.Stderr, "unknown schema: %s\n", e.Schema)
			continue
		}
		proxy := ftm.NewEntityProxy(sc, e.ID)
		for name, vals := range e.Properties {
			_ = proxy.Add(name, vals, false)
		}
		// revalidate and normalize: emit cleaned dict
		_ = sc.Validate(proxy.ToDict()["properties"].(map[string][]string))
		_ = enc.Encode(proxy.ToDict())
	}
}

func pretty() {
	br := bufio.NewScanner(os.Stdin)
	for br.Scan() {
		line := br.Text()
		// best effort to pretty-print a single JSON object per line
		var obj any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			fmt.Println(line) // passthrough
			continue
		}
		buf, _ := json.MarshalIndent(obj, "", "  ")
		os.Stdout.Write(buf)
		os.Stdout.Write([]byte("\n"))
	}
}

func sign() {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	key := fs.String("key", "", "HMAC signature key")
	_ = fs.Parse(os.Args[2:])
	ns := ftm.NewNamespace(*key)
	m := ftm.Default()
	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	for {
		var e entityJSON
		if err := dec.Decode(&e); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "error decoding JSON: %v\n", err)
			os.Exit(1)
		}
		sc := m.Get(e.Schema)
		if sc == nil {
			continue
		}
		proxy := ftm.NewEntityProxy(sc, e.ID)
		for name, vals := range e.Properties {
			_ = proxy.Add(name, vals, true)
		}
		signed := ns.Apply(proxy, false)
		_ = enc.Encode(signed.ToDict())
	}
}
