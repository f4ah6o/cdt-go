package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/f4ah6o/cdt-go/internal/graph"
	"github.com/f4ah6o/cdt-go/internal/render"
)

type Result struct {
	Dir    string
	Output string
}

func GoTest(g *graph.Graph) (*Result, error) {
	dir, err := os.MkdirTemp("", "cdt-test-*")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module cdt/tmp\n\ngo 1.22\n"), 0o644); err != nil {
		return nil, err
	}
	if err := render.Code(g, dir); err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	res := &Result{Dir: dir, Output: out.String()}
	if err != nil {
		return res, fmt.Errorf("go test failed: %w", err)
	}
	return res, nil
}
