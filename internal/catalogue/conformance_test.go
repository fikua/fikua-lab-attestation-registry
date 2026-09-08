package catalogue_test

import (
	"testing"

	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
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

// conformantDefinition returns a minimal Definition that ValidateConformance
// accepts outright, so each negative test below only has to break the one
// rule it's checking.
func conformantDefinition() model.Definition {
	return model.Definition{
		Rulebook: model.AttestationRulebook{
			Category:   model.CategoryNonQualifiedEAA,
			Revocation: model.RevocationNotApplicableShortLived,
			// 24h, the VCR_01b ceiling ValidateConformance enforces.
			MaxValiditySeconds: 24 * 60 * 60,
		},
		Scheme: model.AttestationScheme{
			ID:               "urn:fikua:test:1",
			AttestationLoS:   model.AssuranceModerate,
			BindingType:      model.BindingKey,
			SupportedFormats: []model.CredentialFormat{model.FormatSDJWTVC},
		},
	}
}

func TestValidateConformanceAcceptsAConformantDefinition(t *testing.T) {
	def := conformantDefinition()
	if violations := catalogue.ValidateConformance(def, map[string]int{def.Scheme.ID: 1}); len(violations) != 0 {
		t.Fatalf("expected no violations, got: %v", violations)
	}
}

func TestValidateConformanceRejectsInvalidAssuranceLevel(t *testing.T) {
	def := conformantDefinition()
	def.Scheme.AttestationLoS = "not-a-real-los"
	violations := catalogue.ValidateConformance(def, map[string]int{def.Scheme.ID: 1})
	if !containsSubstring(violations, "attestationLoS") {
		t.Fatalf("expected an attestationLoS violation, got: %v", violations)
	}
}

func TestValidateConformanceRejectsInvalidBindingType(t *testing.T) {
	def := conformantDefinition()
	def.Scheme.BindingType = "not-a-real-binding"
	violations := catalogue.ValidateConformance(def, map[string]int{def.Scheme.ID: 1})
	if !containsSubstring(violations, "bindingType") {
		t.Fatalf("expected a bindingType violation, got: %v", violations)
	}
}

func TestValidateConformanceRejectsInvalidCategory(t *testing.T) {
	def := conformantDefinition()
	def.Rulebook.Category = "not-a-real-category"
	violations := catalogue.ValidateConformance(def, map[string]int{def.Scheme.ID: 1})
	if !containsSubstring(violations, "category") {
		t.Fatalf("expected a category violation, got: %v", violations)
	}
}

func TestValidateConformanceRejectsW3CFormatForNonEAACategory(t *testing.T) {
	def := conformantDefinition()
	def.Rulebook.Category = model.CategoryQEAA
	def.Scheme.SupportedFormats = []model.CredentialFormat{model.CredentialFormat("jwt_vc_json")}
	violations := catalogue.ValidateConformance(def, map[string]int{def.Scheme.ID: 1})
	if !containsSubstring(violations, "W3C VC") {
		t.Fatalf("expected a W3C VC format violation, got: %v", violations)
	}
}

func TestValidateConformanceRejectsDuplicateSchemeID(t *testing.T) {
	def := conformantDefinition()
	violations := catalogue.ValidateConformance(def, map[string]int{def.Scheme.ID: 2})
	if !containsSubstring(violations, "registered more than once") {
		t.Fatalf("expected a duplicate-id violation, got: %v", violations)
	}
}

func TestValidateConformanceRejectsShortLivedWithoutMaxValidity(t *testing.T) {
	def := conformantDefinition()
	def.Rulebook.MaxValiditySeconds = 0
	violations := catalogue.ValidateConformance(def, map[string]int{def.Scheme.ID: 1})
	if !containsSubstring(violations, "maxValiditySeconds is unset") {
		t.Fatalf("expected a missing-maxValiditySeconds violation, got: %v", violations)
	}
}

func TestValidateConformanceRejectsMaxValidityOverTheShortLivedCeiling(t *testing.T) {
	def := conformantDefinition()
	def.Rulebook.MaxValiditySeconds = 25 * 60 * 60 // over the 24h ceiling
	violations := catalogue.ValidateConformance(def, map[string]int{def.Scheme.ID: 1})
	if !containsSubstring(violations, "exceeds the 24h short-lived ceiling") {
		t.Fatalf("expected a maxValiditySeconds-ceiling violation, got: %v", violations)
	}
}

func TestIDCountsTalliesDuplicates(t *testing.T) {
	defs := []model.Definition{
		{Scheme: model.AttestationScheme{ID: "a"}},
		{Scheme: model.AttestationScheme{ID: "a"}},
		{Scheme: model.AttestationScheme{ID: "b"}},
	}
	counts := catalogue.IDCounts(defs)
	if counts["a"] != 2 || counts["b"] != 1 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}
