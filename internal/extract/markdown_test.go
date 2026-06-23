package extract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func TestMarkdownExtractsProjectionFencesAndPreservesOrdinaryFences(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "add.md")
	src := `# Add

Add sums two integers.

` + "```text\nnot projected\n```\n\n```go file=add.go symbol=Add\npackage calc\n\nfunc Add(a int, b int) int { return a + b }\n```\n\n```go file=add_test.go test verifies=Add\npackage calc\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {}\n```\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	b := graph.NewBuilder()
	if err := Markdown(path, b); err != nil {
		t.Fatal(err)
	}

	g := b.Graph()
	var hasCode, hasTest, preservedFence bool
	for _, n := range g.Nodes {
		switch n.Kind {
		case "code":
			hasCode = n.Path == "add.go" && strings.Contains(n.Content, "func Add")
		case "test":
			hasTest = n.Path == "add_test.go" && strings.Contains(n.Content, "func TestAdd")
		case "doc":
			if strings.Contains(n.Content, "```text") && strings.Contains(n.Content, "not projected") {
				preservedFence = true
			}
		}
	}
	if !hasCode || !hasTest || !preservedFence {
		t.Fatalf("hasCode=%v hasTest=%v preservedFence=%v nodes=%#v", hasCode, hasTest, preservedFence, g.Nodes)
	}
}

func TestMarkdownUnclosedFence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.md")
	if err := os.WriteFile(path, []byte("# Bad\n\n```go file=bad.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Markdown(path, graph.NewBuilder())
	if err == nil || !strings.Contains(err.Error(), "unclosed fenced code block") {
		t.Fatalf("Markdown() error = %v, want unclosed fenced code block", err)
	}
}

func TestMarkdownFlushesDocsBeforeSwitchingHeadings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ops.md")
	src := `# Add
Add docs.

# Subtract
Subtract docs.
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	b := graph.NewBuilder()
	if err := Markdown(path, b); err != nil {
		t.Fatal(err)
	}

	described := map[string]string{}
	g := b.Graph()
	for _, e := range g.Edges {
		if e.Kind != "describes" {
			continue
		}
		var doc graph.Node
		for _, n := range g.Nodes {
			if n.ID == e.From {
				doc = n
				break
			}
		}
		described[e.To] = doc.Content
	}
	if !strings.Contains(described[graph.ConceptID("Add")], "Add docs.") {
		t.Fatalf("Add docs not attributed correctly: %#v", described)
	}
	if !strings.Contains(described[graph.ConceptID("Subtract")], "Subtract docs.") {
		t.Fatalf("Subtract docs not attributed correctly: %#v", described)
	}
	if strings.Contains(described[graph.ConceptID("Add")], "Subtract docs.") {
		t.Fatalf("Add doc contains Subtract section: %#v", described)
	}
}
