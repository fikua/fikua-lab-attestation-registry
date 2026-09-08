package catalogue_test

import (
	"testing"

	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
)

// TestAllBundledDefinitionsAreConformant is this registry's publish gate
// (FR-13/FR-14): every bundled definition must pass ARF 3.0 Annex 2 Topic 12
// checks that are verifiable without an external ecosystem-wide registry.
// Fikua has no separate "submit for publication" step — adding a JSON file
// under data/attestations/ and passing this test IS publishing — so this
// test is where the gate lives.
func TestAllBundledDefinitionsAreConformant(t *testing.T) {
	cat := loadBundled(t)
	all := cat.All()
	ids := catalogue.IDCounts(all)

	for _, def := range all {
		if violations := catalogue.ValidateConformance(def, ids); len(violations) != 0 {
			t.Errorf("%s: conformance violations: %v", def.Scheme.ID, violations)
		}
	}
}
