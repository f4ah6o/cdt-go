package extract

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func GoFile(path string, b *graph.Builder) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fileNode := b.AddNode(graph.Node{ID: graph.FileID(path), Kind: "file", Path: slash(path)})
	scanner := bufio.NewScanner(f)
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
				addMarkedBlock(path, fileNode.ID, kind, meta, lines, startLine, lineNo, b)
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

func addMarkedBlock(path, fileNodeID, kind string, meta Meta, lines []string, start, end int, b *graph.Builder) {
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
		n := b.AddNode(graph.Node{ID: graph.NodeID(kind, "go", fmt.Sprintf("%s:%d", path, start)), Kind: kind, Lang: "go", Path: slash(file), Symbol: meta.First("symbol"), Content: content, Source: &graph.Source{Path: slash(path), StartLine: start, EndLine: end}})
		b.AddEdge(graph.Edge{From: n.ID, To: targetFile.ID, Kind: "renders_to"})
		if kind == "code" {
			b.AddEdge(graph.Edge{From: n.ID, To: concept.ID, Kind: "implements"})
		} else {
			b.AddEdge(graph.Edge{From: n.ID, To: concept.ID, Kind: "verifies"})
		}
	}
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
