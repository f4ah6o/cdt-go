package extract

import (
	"io/fs"
	"path/filepath"

	"github.com/f4ah6o/cdt-go/internal/graph"
)

func Paths(paths []string) (*graph.Graph, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	b := graph.NewBuilder()
	for _, root := range paths {
		if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == ".cdt" || name == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			switch filepath.Ext(path) {
			case ".md":
				return Markdown(path, b)
			case ".go":
				return GoFile(path, b)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return b.Graph(), nil
}
