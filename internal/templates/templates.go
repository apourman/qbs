package templates

import (
	"embed"
	"fmt"
	"io/fs"
)

// Files is the canonical AI template filesystem embedded in qbs.
// Consumers should treat returned data as read-only.
//
//go:embed canonical
var Files embed.FS

// Read returns a canonical template by its path beneath canonical/.
func Read(name string) ([]byte, error) {
	b, err := fs.ReadFile(Files, "canonical/"+name)
	if err != nil {
		return nil, fmt.Errorf("read embedded template %q: %w", name, err)
	}
	return b, nil
}
