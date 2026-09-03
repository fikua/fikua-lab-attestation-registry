package catalogue

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
)

// LoadFS loads every *.json attestation definition found directly under dir
// in the given filesystem (typically an embed.FS over data/attestations).
func LoadFS(filesystem fs.FS, dir string) (*Catalogue, error) {
	entries, err := fs.ReadDir(filesystem, dir)
	if err != nil {
		return nil, fmt.Errorf("reading attestation definitions dir %q: %w", dir, err)
	}

	definitions := make([]model.Definition, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := path.Join(dir, entry.Name())
		raw, err := fs.ReadFile(filesystem, path)
		if err != nil {
			return nil, fmt.Errorf("reading %q: %w", path, err)
		}

		var definition model.Definition
		if err := json.Unmarshal(raw, &definition); err != nil {
			return nil, fmt.Errorf("parsing %q: %w", path, err)
		}
		definitions = append(definitions, definition)
	}

	return New(definitions), nil
}
