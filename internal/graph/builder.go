package graph

import (
	"regexp"
	"strings"
)

type Builder struct {
	g     *Graph
	nodes map[string]Node
	edges map[string]Edge
}

func NewBuilder() *Builder {
	return &Builder{
		g:     New(),
		nodes: make(map[string]Node),
		edges: make(map[string]Edge),
	}
}

func (b *Builder) AddNode(n Node) Node {
	if n.ID == "" {
		n.ID = NodeID(n.Kind, n.Lang, firstNonEmpty(n.Symbol, n.Name, n.Path))
	}
	if old, ok := b.nodes[n.ID]; ok {
		return old
	}
	b.nodes[n.ID] = n
	return n
}

func (b *Builder) AddEdge(e Edge) {
	if e.From == "" || e.To == "" || e.Kind == "" {
		return
	}
	key := e.From + "\x00" + e.To + "\x00" + e.Kind
	if _, ok := b.edges[key]; ok {
		return
	}
	b.edges[key] = e
}

func (b *Builder) Graph() *Graph {
	b.g.Nodes = b.g.Nodes[:0]
	b.g.Edges = b.g.Edges[:0]
	for _, n := range b.nodes {
		b.g.Nodes = append(b.g.Nodes, n)
	}
	for _, e := range b.edges {
		b.g.Edges = append(b.g.Edges, e)
	}
	Canonicalize(b.g)
	return b.g
}

func FileID(path string) string {
	return "file:" + sanitize(path)
}

func ConceptID(name string) string {
	return "concept:" + sanitize(name)
}

func NodeID(kind, lang, name string) string {
	parts := []string{kind}
	if lang != "" {
		parts = append(parts, lang)
	}
	parts = append(parts, sanitize(name))
	return strings.Join(parts, ":")
}

var unsafeID = regexp.MustCompile(`[^A-Za-z0-9_.\-/]+`)

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\\", "/")
	s = unsafeID.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-./")
	if s == "" {
		return "unknown"
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return "unknown"
}
