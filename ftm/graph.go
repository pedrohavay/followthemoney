package ftm

// Graph models FtM data as a property graph of nodes and edges.

// Node can represent an entity or a reified property value (e.g., name or url).
type Node struct {
	Type   PropertyType
	Value  string
	ID     string
	Proxy  *EntityProxy
	Schema *Schema
}

func NewNode(t PropertyType, value string, proxy *EntityProxy, schema *Schema) *Node {
	id, ok := t.NodeID(value)
	if !ok {
		id = ""
	}
	sc := schema
	if proxy != nil {
		sc = proxy.Schema
	}
	return &Node{Type: t, Value: value, ID: id, Proxy: proxy, Schema: sc}
}

func NodeFromProxy(e *EntityProxy) *Node {
	if e == nil || e.ID == "" {
		return nil
	}
	return NewNode(registry.Entity, e.ID, e, nil)
}

// Edge links two nodes. If Proxy is set, it encodes an entity-as-edge (relationship schema).
type Edge struct {
	ID       string
	Weight   float64
	SourceID string
	TargetID string
	Prop     *Property
	Proxy    *EntityProxy
	Schema   *Schema
	graph    *Graph
}

func newEdge(g *Graph, src, dst *Node, proxy *EntityProxy, prop *Property, value string) *Edge {
	id := src.ID + "<>" + dst.ID
	e := &Edge{ID: id, SourceID: src.ID, TargetID: dst.ID, Weight: 1.0, Prop: prop, Proxy: proxy, graph: g}
	if prop != nil && value != "" {
		e.Weight = prop.Type.Specificity(value)
	}
	if proxy != nil {
		e.ID = src.ID + "<" + proxy.ID + ">" + dst.ID
		e.Schema = proxy.Schema
	}
	return e
}

func (e *Edge) Source() *Node { return e.graph.nodes[e.SourceID] }
func (e *Edge) Target() *Node { return e.graph.nodes[e.TargetID] }

func (e *Edge) SourceProp() *Property {
	if e.Schema != nil && e.Schema.EdgeSource != "" {
		if p := e.Schema.Get(e.Schema.EdgeSource); p != nil {
			if p.Reverse != nil {
				return p.Reverse
			}
		}
	}
	if e.Prop != nil {
		return e.Prop
	}
	return nil
}

func (e *Edge) TargetProp() *Property {
	if e.Schema != nil && e.Schema.EdgeTarget != "" {
		if p := e.Schema.Get(e.Schema.EdgeTarget); p != nil {
			return p.Reverse
		}
	}
	if e.Prop != nil {
		return e.Prop.Reverse
	}
	return nil
}

func (e *Edge) TypeName() string {
	if e.Schema != nil {
		return e.Schema.Name
	}
	if e.Prop != nil {
		return e.Prop.Name
	}
	return ""
}

// Graph aggregates nodes and edges derived from entities.
type Graph struct {
	edgeTypes []PropertyType
	edges     map[string]*Edge
	nodes     map[string]*Node
	proxies   map[string]*EntityProxy
}

func NewGraph(edgeTypes []PropertyType) *Graph {
	if edgeTypes == nil {
		edgeTypes = []PropertyType{registry.Name, registry.URL, registry.Country}
	}
	g := &Graph{edgeTypes: []PropertyType{}, edges: map[string]*Edge{}, nodes: map[string]*Node{}, proxies: map[string]*EntityProxy{}}
	for _, t := range edgeTypes {
		if t.Matchable() {
			g.edgeTypes = append(g.edgeTypes, t)
		}
	}
	return g
}

func (g *Graph) Flush() {
	g.edges = map[string]*Edge{}
	g.nodes = map[string]*Node{}
	g.proxies = map[string]*EntityProxy{}
}
func (g *Graph) Queue(id string, p *EntityProxy) {
	if _, ok := g.proxies[id]; !ok || p != nil {
		g.proxies[id] = p
	}
}
func (g *Graph) Queued() []string {
	xs := []string{}
	for id, p := range g.proxies {
		if p == nil {
			xs = append(xs, id)
		}
	}
	return xs
}

func (g *Graph) addEdgeProxy(proxy *EntityProxy, source, target string) {
	sp := proxy.Schema.Get(proxy.Schema.EdgeSource)
	tp := proxy.Schema.Get(proxy.Schema.EdgeTarget)
	if sp == nil || tp == nil {
		return
	}
	srcNode := g.getNodeStub(sp, source)
	dstNode := g.getNodeStub(tp, target)
	if srcNode == nil || dstNode == nil || srcNode.ID == "" || dstNode.ID == "" {
		return
	}
	e := newEdge(g, srcNode, dstNode, proxy, nil, "")
	g.edges[e.ID] = e
}

func (g *Graph) getNodeStub(prop *Property, value string) *Node {
	if prop.Type.Name() == registry.Entity.Name() {
		g.Queue(value, nil)
	}
	n := NewNode(prop.Type, value, nil, prop.Range)
	if n.ID == "" {
		return n
	}
	if g.nodes[n.ID] == nil {
		g.nodes[n.ID] = n
	}
	return g.nodes[n.ID]
}

func (g *Graph) addNode(proxy *EntityProxy) {
	ent := NodeFromProxy(proxy)
	if ent == nil || ent.ID == "" {
		return
	}
	g.nodes[ent.ID] = ent
	for name, vals := range proxy.props {
		p := proxy.Schema.Get(name)
		if p == nil {
			continue
		}
		used := false
		for _, t := range g.edgeTypes {
			if p.Type.Name() == t.Name() {
				used = true
				break
			}
		}
		if !used {
			continue
		}
		for _, v := range vals {
			node := g.getNodeStub(p, v)
			if node == nil || node.ID == "" {
				continue
			}
			e := newEdge(g, ent, node, nil, p, v)
			if e.Weight > 0 {
				g.edges[e.ID] = e
			}
		}
	}
}

// Add integrates an entity proxy as either an edge (relationship entity) or node.
func (g *Graph) Add(proxy *EntityProxy) {
	if proxy == nil || proxy.ID == "" {
		return
	}
	g.Queue(proxy.ID, proxy)
	if proxy.Schema.Edge {
		for _, pair := range proxy.EdgePairs() {
			g.addEdgeProxy(proxy, pair[0], pair[1])
		}
	} else {
		g.addNode(proxy)
	}
}

func (g *Graph) Nodes() []*Node {
	out := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, n)
	}
	return out
}
func (g *Graph) Edges() []*Edge {
	out := make([]*Edge, 0, len(g.edges))
	for _, e := range g.edges {
		out = append(out, e)
	}
	return out
}
