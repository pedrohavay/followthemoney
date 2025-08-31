package ftm

// PropertyType defines behavior for a value type.
// Implementations should be stateless and reusable.
type PropertyType interface {
	Name() string
	Group() string   // logical group name, may be empty
	Label() string   // human-readable name
	Matchable() bool // included in matching/comparison
	Pivot() bool     // used to form graph pivots
	MaxLength() int  // maximum length of a single value
	TotalSize() int  // limit to total accumulated string length (0 = unlimited)

	Validate(value string) bool
	Clean(text string, fuzzy bool, format string, proxy *EntityProxy) (string, bool)
	Specificity(value string) float64
	Caption(value string, format string) string
	NodeID(value string) (string, bool)
	CountryHint(value string) (string, bool)
}

// BaseType offers default implementations.
type BaseType struct {
	name      string
	group     string
	label     string
	matchable bool
	pivot     bool
	maxLength int
	totalSize int
}

func (b BaseType) Name() string                          { return b.name }
func (b BaseType) Group() string                         { return b.group }
func (b BaseType) Label() string                         { return b.label }
func (b BaseType) Matchable() bool                       { return b.matchable }
func (b BaseType) Pivot() bool                           { return b.pivot }
func (b BaseType) MaxLength() int                        { return b.maxLength }
func (b BaseType) TotalSize() int                        { return b.totalSize }
func (b BaseType) Specificity(string) float64            { return 0.0 }
func (b BaseType) Caption(value string, _ string) string { return value }
func (b BaseType) NodeID(value string) (string, bool)    { return b.name + ":" + value, true }
func (b BaseType) CountryHint(string) (string, bool)     { return "", false }
