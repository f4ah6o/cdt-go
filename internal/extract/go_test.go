package extract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func TestGoFileMarkersInheritPreamble(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "add.go")
	src := `// Copyright test.
package calc

import "testing"

// cdt:code start concept=Add file=add.go symbol=Add
func Add(a int, b int) int {
	return a + b
}
// cdt:code end

// cdt:test start verifies=Add file=add_test.go
func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Fatal("unexpected result")
	}
}
// cdt:test end
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	b := graph.NewBuilder()
	if err := GoFile(path, b); err != nil {
		t.Fatal(err)
	}

	g := b.Graph()
	code := findNode(t, g, "code")
	if code.Preamble != "package calc\n" {
		t.Fatalf("unexpected code preamble %q", code.Preamble)
	}
	if strings.Contains(code.Content, "package calc") {
		t.Fatalf("fragment content should not include inherited preamble: %q", code.Content)
	}
	test := findNode(t, g, "test")
	if test.Preamble != "package calc\nimport \"testing\"\n" {
		t.Fatalf("unexpected test preamble %q", test.Preamble)
	}
}

func TestGoFileFullFileMarkerDoesNotInheritPreamble(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "add.go")
	src := `package calc

// cdt:code start concept=Add file=add.go
package other

func Add(a int, b int) int { return a + b }
// cdt:code end
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	b := graph.NewBuilder()
	if err := GoFile(path, b); err != nil {
		t.Fatal(err)
	}

	code := findNode(t, b.Graph(), "code")
	if code.Preamble != "" {
		t.Fatalf("full-file block preamble = %q, want empty", code.Preamble)
	}
}

func findNode(t *testing.T, g *graph.Graph, kind string) graph.Node {
	t.Helper()
	for _, n := range g.Nodes {
		if n.Kind == kind {
			return n
		}
	}
	t.Fatalf("missing %s node in %#v", kind, g.Nodes)
	return graph.Node{}
}
