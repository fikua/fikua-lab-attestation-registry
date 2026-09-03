package catalogue_test

import (
	"strings"
	"testing"

	"github.com/fikua/fikua-lab-attestation-registry/data/attestations"
	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
)

func mustSchema(t *testing.T, schemeID string, format model.CredentialFormat) model.FormatSchema {
	t.Helper()
	cat, err := catalogue.LoadFS(attestations.FS, ".")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	definition, err := cat.Get(schemeID)
	if err != nil {
		t.Fatalf("Get(%q): %v", schemeID, err)
	}
	schema := definition.Scheme.SchemaFor(format)
	if schema == nil {
		t.Fatalf("no schema for format %s in %q", format, schemeID)
	}
	return *schema
}

func containsSubstring(violations []string, substr string) bool {
	for _, v := range violations {
		if strings.Contains(v, substr) {
			return true
		}
	}
	return false
}

func TestValidClaimsProduceNoViolations(t *testing.T) {
	schema := mustSchema(t, "urn:eudi:pid:1", model.FormatSDJWTVC)

	claims := map[string]any{
		"family_name":       "Dupont",
		"given_name":        "Jean",
		"birthdate":         "1980-05-23",
		"place_of_birth":    map[string]any{"country": "FR"},
		"nationalities":     []string{"FR"},
		"picture":           "data:image/jpeg;base64,...",
		"issuing_authority": "FR",
		"issuing_country":   "FR",
	}

	if violations := catalogue.Validate(schema, claims); len(violations) != 0 {
		t.Fatalf("unexpected violations: %v", violations)
	}
}

func TestMissingMandatoryClaimIsReported(t *testing.T) {
	schema := mustSchema(t, "urn:eudi:pid:1", model.FormatSDJWTVC)

	claims := map[string]any{"given_name": "Jean"}

	violations := catalogue.Validate(schema, claims)
	if !containsSubstring(violations, "family_name") {
		t.Fatalf("expected a violation mentioning family_name, got %v", violations)
	}
}

func TestUnknownClaimIsReported(t *testing.T) {
	schema := mustSchema(t, "urn:fikua:padro:barcelona:1", model.FormatSDJWTVC)

	claims := map[string]any{"not_a_real_claim": "value"}

	violations := catalogue.Validate(schema, claims)
	if !containsSubstring(violations, "not_a_real_claim") {
		t.Fatalf("expected a violation mentioning not_a_real_claim, got %v", violations)
	}
}
