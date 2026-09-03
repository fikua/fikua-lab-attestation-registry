// Package catalogue loads and serves the Catalogue of Attestations (ARF 3.0 /
// TS11 §4): the in-memory registry of AttestationDefinitions, keyed by
// scheme ID.
package catalogue

import (
	"fmt"

	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
)

// Catalogue is an in-memory, read-only registry of attestation definitions.
type Catalogue struct {
	byID map[string]model.Definition
}

// New builds a Catalogue from a list of definitions, indexed by scheme ID.
func New(definitions []model.Definition) *Catalogue {
	byID := make(map[string]model.Definition, len(definitions))
	for _, d := range definitions {
		byID[d.Scheme.ID] = d
	}
	return &Catalogue{byID: byID}
}

// Get returns the definition registered under schemeID, or an error if none
// exists.
func (c *Catalogue) Get(schemeID string) (model.Definition, error) {
	d, ok := c.byID[schemeID]
	if !ok {
		return model.Definition{}, fmt.Errorf("no attestation definition registered for id %q", schemeID)
	}
	return d, nil
}

// All returns every registered definition, in no particular order.
func (c *Catalogue) All() []model.Definition {
	all := make([]model.Definition, 0, len(c.byID))
	for _, d := range c.byID {
		all = append(all, d)
	}
	return all
}
