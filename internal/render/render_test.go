package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func TestCodeWritesInheritedPreambleOnce(t *testing.T) {
	dir := t.TempDir()
	g := &graph.Graph{
		Version: graph.Version,
		Nodes: []graph.Node{
			{ID: "code:add", Kind: "code", Lang: "go", Path: "calc/add.go", Preamble: "package calc\n\nimport \"testing\"\n", Content: "func Add(a int, b int) int { return a + b }\n"},
			{ID: "code:sub", Kind: "code", Lang: "go", Path: "calc/add.go", Preamble: "package calc\n\nimport \"testing\"\n", Content: "func Sub(a int, b int) int { return a - b }\n"},
		},
	}

	if err := Code(g, dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "calc", "add.go"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Count(got, "package calc") != 1 {
		t.Fatalf("package clause count = %d in:\n%s", strings.Count(got, "package calc"), got)
	}
	if !strings.Contains(got, "func Add") || !strings.Contains(got, "func Sub") {
		t.Fatalf("missing rendered content:\n%s", got)
	}
}

func TestCodeRejectsUnsafePaths(t *testing.T) {
	for _, p := range []string{"../evil.go", "/tmp/evil.go"} {
		t.Run(p, func(t *testing.T) {
			g := &graph.Graph{
				Version: graph.Version,
				Nodes:   []graph.Node{{ID: "code:evil", Kind: "code", Lang: "go", Path: p, Content: "package evil\n"}},
			}
			if err := Code(g, t.TempDir()); err == nil {
				t.Fatalf("Code() error = nil, want unsafe path error")
			}
		})
	}
}

func TestCodeDoesNotRenderTests(t *testing.T) {
	dir := t.TempDir()
	g := &graph.Graph{
		Version: graph.Version,
		Nodes: []graph.Node{
			{ID: "code:add", Kind: "code", Lang: "go", Path: "add.go", Content: "package calc\n\nfunc Add() {}\n"},
			{ID: "test:add", Kind: "test", Lang: "go", Path: "add_test.go", Content: "package calc\n\nfunc TestAdd() {}\n"},
		},
	}

	if err := Code(g, dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "add.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "add_test.go")); !os.IsNotExist(err) {
		t.Fatalf("Code rendered test file, stat err=%v", err)
	}
}

func TestAllGoRendersCodeAndTests(t *testing.T) {
	dir := t.TempDir()
	g := &graph.Graph{
		Version: graph.Version,
		Nodes: []graph.Node{
			{ID: "code:add", Kind: "code", Lang: "go", Path: "add.go", Content: "package calc\n\nfunc Add() {}\n"},
			{ID: "test:add", Kind: "test", Lang: "go", Path: "add_test.go", Content: "package calc\n\nfunc TestAdd() {}\n"},
		},
	}

	if err := AllGo(g, dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"add.go", "add_test.go"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
}

func TestCodePreservesSourceOrderWithinPath(t *testing.T) {
	dir := t.TempDir()
	g := &graph.Graph{
		Version: graph.Version,
		Nodes: []graph.Node{
			{ID: "code:go:sample.go:10", Kind: "code", Lang: "go", Path: "sample.go", Content: "func Later() {}\n", Source: &graph.Source{Path: "spec.md", StartLine: 10, EndLine: 12}},
			{ID: "code:go:sample.go:2", Kind: "code", Lang: "go", Path: "sample.go", Content: "package calc\n", Source: &graph.Source{Path: "spec.md", StartLine: 2, EndLine: 4}},
		},
	}

	if err := Code(g, dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "sample.go"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if strings.Index(got, "package calc") > strings.Index(got, "func Later") {
		t.Fatalf("rendered out of source order:\n%s", got)
	}
}
