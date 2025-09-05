package ftmschema

import "embed"

// Files embeds the YAML schema definitions shipped with the library.
//go:embed *.yaml *.yml
var Files embed.FS

