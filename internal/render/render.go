package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func Code(g *graph.Graph, outDir string) error {
	return writeNodes(g, outDir, map[string]bool{"code": true, "test": true})
}

func Tests(g *graph.Graph, outDir string) error {
	return writeNodes(g, outDir, map[string]bool{"test": true})
}

func writeNodes(g *graph.Graph, outDir string, kinds map[string]bool) error {
	byPath := map[string][]graph.Node{}
	for _, n := range g.Nodes {
		if !kinds[n.Kind] || n.Path == "" {
			continue
		}
		byPath[n.Path] = append(byPath[n.Path], n)
	}
	paths := make([]string, 0, len(byPath))
	for p := range byPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		var b strings.Builder
		for i, n := range byPath[p] {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(strings.TrimRight(n.Content, "\n"))
			b.WriteString("\n")
		}
		target := filepath.Join(outDir, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(b.String()), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func Docs(g *graph.Graph, outDir string) error {
	var b strings.Builder
	b.WriteString("# cdt generated docs\n\n")
	for _, n := range g.Nodes {
		if n.Kind == "doc" && strings.TrimSpace(n.Content) != "" {
			b.WriteString(strings.TrimSpace(n.Content))
			b.WriteString("\n\n")
		}
	}
	for _, n := range g.Nodes {
		if n.Kind != "code" && n.Kind != "test" {
			continue
		}
		info := fmt.Sprintf("go file=%s", n.Path)
		if n.Kind == "test" {
			info += " test"
		}
		if n.Symbol != "" {
			info += " symbol=" + n.Symbol
		}
		b.WriteString("```")
		b.WriteString(info)
		b.WriteString("\n")
		b.WriteString(strings.TrimRight(n.Content, "\n"))
		b.WriteString("\n```\n\n")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "cdt.generated.md"), []byte(b.String()), 0o644)
}
