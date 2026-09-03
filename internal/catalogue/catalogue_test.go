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

func TestLoadsAllTwoBundledDefinitions(t *testing.T) {
	cat := loadBundled(t)
	if got := len(cat.All()); got != 2 {
		t.Fatalf("got %d definitions, want 2", got)
	}
}

// TestPidSchemeHasBothFormats verifies the PID is a single attestation
// definition with two FormatSchemas, per TS11 §4.3.1 (one SchemaMeta per
// attestation type, with a Schema entry per supported format) — not two
// separate scheme ids as this repo initially, incorrectly, modeled it.
func TestPidSchemeHasBothFormats(t *testing.T) {
	cat := loadBundled(t)

	pid, err := cat.Get("urn:eudi:pid:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got := pid.Scheme.SupportedFormats; len(got) != 2 {
		t.Fatalf("supportedFormats = %v, want 2 entries", got)
	}

	sdJwtSchema := pid.Scheme.SchemaFor(model.FormatSDJWTVC)
	if sdJwtSchema == nil {
		t.Fatal("no SD-JWT VC schema found")
	}
	if sdJwtSchema.TypeIdentifier != "urn:eudi:pid:1" {
		t.Errorf("sd-jwt typeIdentifier = %q, want urn:eudi:pid:1", sdJwtSchema.TypeIdentifier)
	}

	familyName := sdJwtSchema.Claim("family_name")
	if familyName == nil {
		t.Fatal("family_name claim not found in sd-jwt schema")
	}
	if familyName.Presence != model.PresenceMandatory {
		t.Errorf("family_name presence = %s, want mandatory", familyName.Presence)
	}
	if familyName.Path[0] != "family_name" {
		t.Errorf("family_name path = %v", familyName.Path)
	}

	mdocSchema := pid.Scheme.SchemaFor(model.FormatMDoc)
	if mdocSchema == nil {
		t.Fatal("no mdoc schema found")
	}
	if mdocSchema.TypeIdentifier != "eu.europa.ec.eudi.pid.1" {
		t.Errorf("mdoc typeIdentifier = %q, want eu.europa.ec.eudi.pid.1", mdocSchema.TypeIdentifier)
	}

	birthDate := mdocSchema.Claim("birth_date")
	if birthDate == nil {
		t.Fatal("birth_date claim not found in mdoc schema")
	}
	if birthDate.Namespace != "eu.europa.ec.eudi.pid.1" {
		t.Errorf("birth_date namespace = %q", birthDate.Namespace)
	}
}

// TestPidMdocTypeIdentifierIsNotASeparateSchemeID guards against
// regressing back to two scheme ids for the same attestation type: the
// mdoc doctype must NOT be independently reachable via cat.Get.
func TestPidMdocTypeIdentifierIsNotASeparateSchemeID(t *testing.T) {
	cat := loadBundled(t)
	if _, err := cat.Get("eu.europa.ec.eudi.pid.1"); err == nil {
		t.Fatal("eu.europa.ec.eudi.pid.1 should not be a top-level scheme id; it's the mdoc FormatSchema.typeIdentifier under urn:eudi:pid:1")
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

// TestEveryDefinitionDeclaresRevocation guards Rulebook template §6
// (SHALL): every Rulebook must state whether attestations are short-lived
// enough that revocation is never necessary, or which mechanism applies.
func TestEveryDefinitionDeclaresRevocation(t *testing.T) {
	cat := loadBundled(t)
	for _, def := range cat.All() {
		if def.Rulebook.Revocation == "" {
			t.Errorf("%s: Rulebook.Revocation must be set (template §6 SHALL)", def.Scheme.ID)
		}
	}
}

// TestEveryDefinitionHasACatalogueUUID guards TS11 §4.3.1: SchemaMeta.id is
// an opaque UUID assigned by the catalogue provider, distinct from this
// registry's own human-readable lookup id.
func TestEveryDefinitionHasACatalogueUUID(t *testing.T) {
	cat := loadBundled(t)
	for _, def := range cat.All() {
		if def.Scheme.CatalogueID == "" {
			t.Errorf("%s: Scheme.CatalogueID must be set (TS11 §4.3.1 SchemaMeta.id)", def.Scheme.ID)
		}
	}
}

func TestAttestationLegalCategoryIsConstrainedToPID(t *testing.T) {
	cat := loadBundled(t)
	pid, err := cat.Get("urn:eudi:pid:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	for _, format := range []model.CredentialFormat{model.FormatSDJWTVC, model.FormatMDoc} {
		schema := pid.Scheme.SchemaFor(format)
		claim := schema.Claim("attestation_legal_category")
		if claim == nil {
			t.Fatalf("%s: attestation_legal_category claim not found", format)
		}
		if len(claim.Enum) != 1 || claim.Enum[0] != "PID" {
			t.Errorf("%s: attestation_legal_category enum = %v, want [PID] (Rulebook §2.6)", format, claim.Enum)
		}
	}
}
