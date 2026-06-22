package extract

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func Markdown(path string, b *graph.Builder) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	docFile := b.AddNode(graph.Node{ID: graph.FileID(path), Kind: "file", Path: slash(path)})

	scanner := bufio.NewScanner(f)
	lineNo := 0
	currentConcept := conceptFromFilename(path)
	var docLines []string
	var docStart int
	inFence := false
	fenceLang := ""
	fenceMeta := ""
	fenceStart := 0
	var fenceLines []string

	flushDoc := func(end int) {
		content := strings.TrimSpace(strings.Join(docLines, "\n"))
		if content == "" {
			docLines = nil
			return
		}
		concept := b.AddNode(graph.Node{ID: graph.ConceptID(currentConcept), Kind: "concept", Name: currentConcept})
		doc := b.AddNode(graph.Node{
			ID:      graph.NodeID("doc", "md", fmt.Sprintf("%s:%d", path, docStart)),
			Kind:    "doc",
			Lang:    "md",
			Path:    slash(path),
			Content: content,
			Source:  &graph.Source{Path: slash(path), StartLine: docStart, EndLine: end},
		})
		b.AddEdge(graph.Edge{From: docFile.ID, To: doc.ID, Kind: "contains"})
		b.AddEdge(graph.Edge{From: doc.ID, To: concept.ID, Kind: "describes"})
		docLines = nil
	}

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			if !inFence {
				flushDoc(lineNo - 1)
				inFence = true
				fenceStart = lineNo
				rest := strings.TrimSpace(strings.TrimPrefix(trim, "```"))
				parts := strings.Fields(rest)
				if len(parts) > 0 {
					fenceLang = parts[0]
					fenceMeta = strings.TrimSpace(strings.TrimPrefix(rest, fenceLang))
				} else {
					fenceLang = ""
					fenceMeta = ""
				}
				fenceLines = nil
				continue
			}
			if err := addFence(path, currentConcept, fenceLang, fenceMeta, fenceLines, fenceStart, lineNo, b); err != nil {
				return err
			}
			inFence = false
			continue
		}
		if inFence {
			fenceLines = append(fenceLines, line)
			continue
		}
		if strings.HasPrefix(trim, "#") {
			name := strings.TrimSpace(strings.TrimLeft(trim, "#"))
			if name != "" {
				currentConcept = name
			}
		}
		if len(docLines) == 0 {
			docStart = lineNo
		}
		docLines = append(docLines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if inFence {
		return fmt.Errorf("%s:%d: unclosed fenced code block", path, fenceStart)
	}
	flushDoc(lineNo)
	return nil
}

func addFence(path, currentConcept, lang, metaText string, lines []string, start, end int, b *graph.Builder) error {
	meta := ParseMeta(metaText)
	if lang != "go" {
		return nil
	}
	file := meta.First("file", "path")
	if file == "" {
		return nil
	}
	conceptName := meta.First("concept", "verifies", "symbol")
	if conceptName == "" {
		conceptName = currentConcept
	}
	content := strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
	concept := b.AddNode(graph.Node{ID: graph.ConceptID(conceptName), Kind: "concept", Name: conceptName})
	targetFile := b.AddNode(graph.Node{ID: graph.FileID(file), Kind: "file", Path: slash(file)})
	kind := "code"
	edgeKind := "implements"
	if meta.Bool("test") || strings.HasSuffix(file, "_test.go") {
		kind = "test"
		edgeKind = "verifies"
	}
	n := b.AddNode(graph.Node{
		ID:      graph.NodeID(kind, lang, fmt.Sprintf("%s:%d", path, start)),
		Kind:    kind,
		Lang:    lang,
		Path:    slash(file),
		Symbol:  meta.First("symbol"),
		Content: content,
		Source:  &graph.Source{Path: slash(path), StartLine: start, EndLine: end},
	})
	b.AddEdge(graph.Edge{From: n.ID, To: targetFile.ID, Kind: "renders_to"})
	b.AddEdge(graph.Edge{From: n.ID, To: concept.ID, Kind: edgeKind})
	return nil
}

func conceptFromFilename(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func slash(path string) string {
	return filepath.ToSlash(path)
}
