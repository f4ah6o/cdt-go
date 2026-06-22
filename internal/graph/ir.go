package graph

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

const Version = "0.1"

type Graph struct {
	Version string `json:"version"`
	Nodes   []Node `json:"nodes"`
	Edges   []Edge `json:"edges"`
}

type Node struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Name    string `json:"name,omitempty"`
	Path    string `json:"path,omitempty"`
	Lang    string `json:"lang,omitempty"`
	Symbol  string `json:"symbol,omitempty"`
	Content string `json:"content,omitempty"`

	Source     *Source `json:"source,omitempty"`
	Inferred   bool    `json:"inferred,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

type Source struct {
	Path      string `json:"path"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`

	Inferred   bool    `json:"inferred,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

func New() *Graph {
	return &Graph{Version: Version}
}

func Read(r io.Reader) (*Graph, error) {
	var g Graph
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&g); err != nil {
		return nil, err
	}
	if g.Version == "" {
		g.Version = Version
	}
	return &g, nil
}

func Write(w io.Writer, g *Graph) error {
	Canonicalize(g)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(g)
}

func Canonicalize(g *Graph) {
	if g.Version == "" {
		g.Version = Version
	}
	sort.Slice(g.Nodes, func(i, j int) bool {
		return g.Nodes[i].ID < g.Nodes[j].ID
	})
	sort.Slice(g.Edges, func(i, j int) bool {
		a := fmt.Sprintf("%s\x00%s\x00%s", g.Edges[i].From, g.Edges[i].To, g.Edges[i].Kind)
		b := fmt.Sprintf("%s\x00%s\x00%s", g.Edges[j].From, g.Edges[j].To, g.Edges[j].Kind)
		return a < b
	})
}
