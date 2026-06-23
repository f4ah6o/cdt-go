package runner

import (
	"os"
	"testing"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func TestGoTest(t *testing.T) {
	g := &graph.Graph{
		Version: graph.Version,
		Nodes: []graph.Node{
			{ID: "concept:Add", Kind: "concept", Name: "Add"},
			{ID: "code:add", Kind: "code", Lang: "go", Path: "add.go", Content: "package calc\n\nfunc Add(a int, b int) int { return a + b }\n"},
			{ID: "test:add", Kind: "test", Lang: "go", Path: "add_test.go", Content: "package calc\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) { if Add(1, 2) != 3 { t.Fatal(\"unexpected result\") } }\n"},
		},
	}
	res, err := GoTest(g)
	if err != nil {
		t.Fatalf("%v\n%s", err, res.Output)
	}
	defer os.RemoveAll(res.Dir)
}
