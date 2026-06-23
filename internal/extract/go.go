package extract

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func GoFile(path string, b *graph.Builder) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	preamble := goPreamble(content)

	fileNode := b.AddNode(graph.Node{ID: graph.FileID(path), Kind: "file", Path: slash(path)})
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	in := false
	kind := ""
	meta := Meta{}
	startLine := 0
	var lines []string

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		marker := markerText(line)
		if marker != "" && strings.HasPrefix(marker, "cdt:") {
			rest := strings.TrimPrefix(marker, "cdt:")
			fields := strings.Fields(rest)
			if len(fields) >= 2 && fields[1] == "start" {
				if in {
					return fmt.Errorf("%s:%d: nested cdt marker", path, lineNo)
				}
				in = true
				kind = fields[0]
				meta = ParseMeta(strings.Join(fields[2:], " "))
				startLine = lineNo
				lines = nil
				continue
			}
			if len(fields) >= 2 && fields[1] == "end" {
				if !in || fields[0] != kind {
					return fmt.Errorf("%s:%d: unmatched cdt marker", path, lineNo)
				}
				addMarkedBlock(path, fileNode.ID, kind, meta, lines, preamble, startLine, lineNo, b)
				in = false
				continue
			}
		}
		if in {
			if kind == "doc" {
				lines = append(lines, stripLineComment(line))
			} else {
				lines = append(lines, line)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if in {
		return fmt.Errorf("%s:%d: unclosed cdt marker", path, startLine)
	}
	return nil
}

func addMarkedBlock(path, fileNodeID, kind string, meta Meta, lines []string, preamble string, start, end int, b *graph.Builder) {
	conceptName := meta.First("concept", "verifies", "symbol")
	if conceptName == "" {
		conceptName = conceptFromFilename(path)
	}
	concept := b.AddNode(graph.Node{ID: graph.ConceptID(conceptName), Kind: "concept", Name: conceptName})
	content := strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"

	switch kind {
	case "doc":
		doc := b.AddNode(graph.Node{ID: graph.NodeID("doc", "md", fmt.Sprintf("%s:%d", path, start)), Kind: "doc", Lang: "md", Path: slash(path), Content: strings.TrimSpace(content), Source: &graph.Source{Path: slash(path), StartLine: start, EndLine: end}})
		b.AddEdge(graph.Edge{From: fileNodeID, To: doc.ID, Kind: "contains"})
		b.AddEdge(graph.Edge{From: doc.ID, To: concept.ID, Kind: "describes"})
	case "code", "test":
		file := meta.First("file", "path")
		if file == "" {
			file = path
		}
		targetFile := b.AddNode(graph.Node{ID: graph.FileID(file), Kind: "file", Path: slash(file)})
		nodePreamble := ""
		if !hasPackageClause(content) {
			nodePreamble = preambleForContent(preamble, content)
		}
		n := b.AddNode(graph.Node{ID: graph.NodeID(kind, "go", fmt.Sprintf("%s:%d", path, start)), Kind: kind, Lang: "go", Path: slash(file), Symbol: meta.First("symbol"), Content: content, Preamble: nodePreamble, Source: &graph.Source{Path: slash(path), StartLine: start, EndLine: end}})
		b.AddEdge(graph.Edge{From: n.ID, To: targetFile.ID, Kind: "renders_to"})
		if kind == "code" {
			b.AddEdge(graph.Edge{From: n.ID, To: concept.ID, Kind: "implements"})
		} else {
			b.AddEdge(graph.Edge{From: n.ID, To: concept.ID, Kind: "verifies"})
		}
	}
}

func goPreamble(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	inImportBlock := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" && len(out) == 0 {
			continue
		}
		if strings.HasPrefix(trim, "//") && len(out) == 0 {
			continue
		}
		if strings.HasPrefix(trim, "package ") {
			out = append(out, line)
			continue
		}
		if strings.HasPrefix(trim, "import (") {
			inImportBlock = true
			out = append(out, line)
			continue
		}
		if inImportBlock {
			out = append(out, line)
			if trim == ")" {
				inImportBlock = false
			}
			continue
		}
		if strings.HasPrefix(trim, "import ") {
			out = append(out, line)
			continue
		}
		if trim == "" && len(out) > 0 {
			continue
		}
		break
	}
	if len(out) == 0 {
		return ""
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}

func hasPackageClause(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "//") {
			continue
		}
		return strings.HasPrefix(trim, "package ")
	}
	return false
}

func preambleForContent(preamble, content string) string {
	if preamble == "" {
		return ""
	}
	lines := strings.Split(strings.TrimRight(preamble, "\n"), "\n")
	if len(lines) == 0 {
		return ""
	}
	out := []string{lines[0]}
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if strings.HasPrefix(trim, "import (") {
			var kept []string
			for i++; i < len(lines); i++ {
				blockLine := lines[i]
				blockTrim := strings.TrimSpace(blockLine)
				if blockTrim == ")" {
					break
				}
				if importUsed(blockTrim, content) {
					kept = append(kept, blockLine)
				}
			}
			if len(kept) > 0 {
				out = append(out, "import (")
				out = append(out, kept...)
				out = append(out, ")")
			}
			continue
		}
		if strings.HasPrefix(trim, "import ") && importUsed(strings.TrimSpace(strings.TrimPrefix(trim, "import ")), content) {
			out = append(out, line)
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n") + "\n"
}

func importUsed(importSpec, content string) bool {
	fields := strings.Fields(importSpec)
	if len(fields) == 0 {
		return false
	}
	if fields[0] == "_" || fields[0] == "." {
		return true
	}
	if len(fields) > 1 {
		return strings.Contains(content, fields[0]+".")
	}
	path := strings.Trim(fields[0], `"`)
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, "-go")
	return strings.Contains(content, name+".")
}

func markerText(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "//")
	return strings.TrimSpace(line)
}

func stripLineComment(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "//") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "//"))
	}
	return line
}
