package catalogue_test

import (
	"testing"

	"github.com/fikua/fikua-lab-attestation-registry/data/attestations"
	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
)

func loadBundled(t *testing.T) *catalogue.Catalogue {
	t.Helper()
	cat, err := catalogue.LoadFS(attestations.FS, ".")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	return cat
}

func TestLoadsAllThreeBundledDefinitions(t *testing.T) {
	cat := loadBundled(t)
	if got := len(cat.All()); got != 3 {
		t.Fatalf("got %d definitions, want 3", got)
	}
}

func TestPidSDJWTSchemeUsesEudiPidVct(t *testing.T) {
	cat := loadBundled(t)

	pid, err := cat.Get("urn:eudi:pid:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got := pid.Scheme.SupportedFormats; len(got) != 1 || got[0] != model.FormatSDJWTVC {
		t.Fatalf("supportedFormats = %v, want [%s]", got, model.FormatSDJWTVC)
	}

	schema := pid.Scheme.SchemaFor(model.FormatSDJWTVC)
	if schema == nil {
		t.Fatal("no SD-JWT VC schema found")
	}

	familyName := schema.Claim("family_name")
	if familyName == nil {
		t.Fatal("family_name claim not found")
	}
	if familyName.Presence != model.PresenceMandatory {
		t.Errorf("family_name presence = %s, want mandatory", familyName.Presence)
	}
	if familyName.Path[0] != "family_name" {
		t.Errorf("family_name path = %v", familyName.Path)
	}
}

func TestPidMdocSchemeUsesEudiPidDoctype(t *testing.T) {
	cat := loadBundled(t)

	pid, err := cat.Get("eu.europa.ec.eudi.pid.1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	schema := pid.Scheme.SchemaFor(model.FormatMDoc)
	if schema == nil {
		t.Fatal("no mdoc schema found")
	}

	birthDate := schema.Claim("birth_date")
	if birthDate == nil {
		t.Fatal("birth_date claim not found")
	}
	if birthDate.Namespace != "eu.europa.ec.eudi.pid.1" {
		t.Errorf("birth_date namespace = %q", birthDate.Namespace)
	}
}

func TestPadroSchemeIsNonQualifiedEaaBoundToPid(t *testing.T) {
	cat := loadBundled(t)

	padro, err := cat.Get("urn:fikua:padro:barcelona:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if padro.Rulebook.Category != model.CategoryNonQualifiedEAA {
		t.Errorf("category = %q, want %q", padro.Rulebook.Category, model.CategoryNonQualifiedEAA)
	}

	schema := padro.Scheme.SchemaFor(model.FormatSDJWTVC)
	binding := schema.Claim("cryptographically_bound_to")
	if binding == nil {
		t.Fatal("cryptographically_bound_to claim not found")
	}
	if binding.Presence != model.PresenceMandatory {
		t.Errorf("cryptographically_bound_to presence = %s, want mandatory", binding.Presence)
	}
}

func TestUnknownSchemeIDReturnsError(t *testing.T) {
	cat := loadBundled(t)
	if _, err := cat.Get("does-not-exist"); err == nil {
		t.Fatal("expected error for unknown scheme id")
	}
}
