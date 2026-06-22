package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/f4ah6o/cdt-go/internal/extract"
	"github.com/f4ah6o/cdt-go/internal/graph"
	"github.com/f4ah6o/cdt-go/internal/render"
	"github.com/f4ah6o/cdt-go/internal/runner"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "graph":
		err = runGraph(os.Args[2:])
	case "check":
		err = runCheck(os.Args[2:])
	case "render":
		err = runRender(os.Args[2:])
	case "test":
		err = runTest(os.Args[2:])
	case "coverage":
		err = runCoverage(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "cdt:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: cdt <graph|check|render|test|coverage> [options]")
}

func runGraph(args []string) error {
	out, paths, err := parseOutputFlag(args, ".cdt/graph.json")
	if err != nil {
		return err
	}
	g, err := extract.Paths(paths)
	if err != nil {
		return err
	}
	if err := graph.Validate(g); err != nil {
		return err
	}
	if out == "-" {
		return graph.Write(os.Stdout, g)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	return graph.Write(f, g)
}

func parseOutputFlag(args []string, defaultOut string) (string, []string, error) {
	out := defaultOut
	paths := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-o" || arg == "--output":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("%s requires a value", arg)
			}
			out = args[i+1]
			i++
		case strings.HasPrefix(arg, "-o="):
			out = strings.TrimPrefix(arg, "-o=")
		case strings.HasPrefix(arg, "--output="):
			out = strings.TrimPrefix(arg, "--output=")
		default:
			paths = append(paths, arg)
		}
	}
	return out, paths, nil
}

func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	in := fs.String("graph", ".cdt/graph.json", "graph path")
	strict := fs.Bool("strict", false, "fail when any concept lacks docs, code, or tests")
	if err := fs.Parse(args); err != nil {
		return err
	}
	g, err := readGraph(*in)
	if err != nil {
		return err
	}
	if err := graph.Validate(g); err != nil {
		return err
	}
	if *strict {
		for _, row := range graph.Coverage(g) {
			if row.Status != "ok" {
				return fmt.Errorf("concept %s is incomplete: docs=%d code=%d tests=%d", row.Concept, row.Docs, row.Code, row.Tests)
			}
		}
	}
	fmt.Println("ok")
	return nil
}

func runRender(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("render target required: docs, code, or test")
	}
	target := args[0]
	fs := flag.NewFlagSet("render "+target, flag.ContinueOnError)
	in := fs.String("graph", ".cdt/graph.json", "graph path")
	out := fs.String("o", ".", "output directory")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	g, err := readGraph(*in)
	if err != nil {
		return err
	}
	switch target {
	case "docs":
		return render.Docs(g, *out)
	case "code":
		return render.Code(g, *out)
	case "test", "tests":
		return render.Tests(g, *out)
	default:
		return fmt.Errorf("unknown render target %q", target)
	}
}

func runTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	in := fs.String("graph", ".cdt/graph.json", "graph path")
	keep := fs.Bool("keep", false, "keep temporary test workspace")
	if err := fs.Parse(args); err != nil {
		return err
	}
	g, err := readGraph(*in)
	if err != nil {
		return err
	}
	res, err := runner.GoTest(g)
	if res != nil && res.Output != "" {
		fmt.Print(res.Output)
	}
	if res != nil && *keep {
		fmt.Println("workspace:", res.Dir)
	} else if res != nil {
		_ = os.RemoveAll(res.Dir)
	}
	return err
}

func runCoverage(args []string) error {
	fs := flag.NewFlagSet("coverage", flag.ContinueOnError)
	in := fs.String("graph", ".cdt/graph.json", "graph path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	g, err := readGraph(*in)
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CONCEPT\tDOCS\tCODE\tTESTS\tSTATUS")
	for _, row := range graph.Coverage(g) {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\n", row.Concept, row.Docs, row.Code, row.Tests, row.Status)
	}
	return w.Flush()
}

func readGraph(path string) (*graph.Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return graph.Read(f)
}
