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
	return writeNodes(g, outDir, map[string]bool{"code": true})
}

func Tests(g *graph.Graph, outDir string) error {
	return writeNodes(g, outDir, map[string]bool{"test": true})
}

func AllGo(g *graph.Graph, outDir string) error {
	return writeNodes(g, outDir, map[string]bool{"code": true, "test": true})
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
		sort.SliceStable(byPath[p], func(i, j int) bool {
			a := byPath[p][i]
			b := byPath[p][j]
			if a.Source != nil && b.Source != nil {
				if a.Source.Path != b.Source.Path {
					return a.Source.Path < b.Source.Path
				}
				if a.Source.StartLine != b.Source.StartLine {
					return a.Source.StartLine < b.Source.StartLine
				}
			} else if a.Source != nil {
				return true
			} else if b.Source != nil {
				return false
			}
			return a.ID < b.ID
		})
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		var b strings.Builder
		preamble := ""
		for i, n := range byPath[p] {
			if n.Preamble != "" {
				if preamble != "" && preamble != n.Preamble {
					return fmt.Errorf("%s: conflicting preambles", p)
				}
				preamble = n.Preamble
			}
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(strings.TrimRight(n.Content, "\n"))
			b.WriteString("\n")
		}
		var content strings.Builder
		if preamble != "" {
			content.WriteString(strings.TrimRight(preamble, "\n"))
			content.WriteString("\n\n")
		}
		content.WriteString(b.String())
		target, err := safeJoin(outDir, p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(content.String()), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func safeJoin(outDir, graphPath string) (string, error) {
	if graphPath == "" {
		return "", fmt.Errorf("empty output path")
	}
	if filepath.IsAbs(graphPath) || filepath.IsAbs(filepath.FromSlash(graphPath)) {
		return "", fmt.Errorf("unsafe output path %q", graphPath)
	}
	clean := filepath.Clean(filepath.FromSlash(graphPath))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe output path %q", graphPath)
	}
	return filepath.Join(outDir, clean), nil
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
		if n.Preamble != "" {
			b.WriteString(strings.TrimRight(n.Preamble, "\n"))
			b.WriteString("\n\n")
		}
		b.WriteString(strings.TrimRight(n.Content, "\n"))
		b.WriteString("\n```\n\n")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "cdt.generated.md"), []byte(b.String()), 0o644)
}
