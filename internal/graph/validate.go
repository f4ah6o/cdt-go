package graph

import "fmt"

var nodeKinds = map[string]bool{
	"file":    true,
	"doc":     true,
	"code":    true,
	"test":    true,
	"concept": true,
}

var edgeKinds = map[string]bool{
	"contains":     true,
	"describes":    true,
	"implements":   true,
	"verifies":     true,
	"renders_to":   true,
	"derived_from": true,
}

func Validate(g *Graph) error {
	ids := make(map[string]string, len(g.Nodes))
	for _, n := range g.Nodes {
		if n.ID == "" {
			return fmt.Errorf("node id is empty")
		}
		if n.Kind == "" {
			return fmt.Errorf("node %s kind is empty", n.ID)
		}
		if !nodeKinds[n.Kind] {
			return fmt.Errorf("node %s has unsupported kind %q", n.ID, n.Kind)
		}
		if oldKind, exists := ids[n.ID]; exists {
			return fmt.Errorf("duplicate node id %q (kinds %s and %s)", n.ID, oldKind, n.Kind)
		}
		ids[n.ID] = n.Kind
	}
	for _, e := range g.Edges {
		if e.From == "" || e.To == "" || e.Kind == "" {
			return fmt.Errorf("edge has empty from/to/kind: %+v", e)
		}
		if !edgeKinds[e.Kind] {
			return fmt.Errorf("edge %s -> %s has unsupported kind %q", e.From, e.To, e.Kind)
		}
		if _, ok := ids[e.From]; !ok {
			return fmt.Errorf("edge references missing from node %q", e.From)
		}
		if _, ok := ids[e.To]; !ok {
			return fmt.Errorf("edge references missing to node %q", e.To)
		}
	}
	return nil
}

type CoverageRow struct {
	Concept string
	Docs    int
	Code    int
	Tests   int
	Status  string
}

func Coverage(g *Graph) []CoverageRow {
	rows := map[string]*CoverageRow{}
	conceptName := map[string]string{}
	for _, n := range g.Nodes {
		if n.Kind == "concept" {
			name := n.Name
			if name == "" {
				name = n.ID
			}
			conceptName[n.ID] = name
			rows[n.ID] = &CoverageRow{Concept: name}
		}
	}
	for _, e := range g.Edges {
		row, ok := rows[e.To]
		if !ok {
			continue
		}
		switch e.Kind {
		case "describes":
			row.Docs++
		case "implements":
			row.Code++
		case "verifies":
			row.Tests++
		}
	}
	out := make([]CoverageRow, 0, len(rows))
	for id, row := range rows {
		if row.Concept == "" {
			row.Concept = conceptName[id]
		}
		row.Status = "ok"
		if row.Docs == 0 || row.Code == 0 || row.Tests == 0 {
			row.Status = "incomplete"
		}
		out = append(out, *row)
	}
	return out
}
