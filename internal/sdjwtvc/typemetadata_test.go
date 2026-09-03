package sdjwtvc_test

import (
	"testing"

	"github.com/fikua/fikua-lab-attestation-registry/data/attestations"
	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
	"github.com/fikua/fikua-lab-attestation-registry/internal/sdjwtvc"
)

func loadBundled(t *testing.T) *catalogue.Catalogue {
	t.Helper()
	cat, err := catalogue.LoadFS(attestations.FS, ".")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	return cat
}

func TestFromSchemeProducesValidTypeMetadataForPidSdJwt(t *testing.T) {
	cat := loadBundled(t)
	def, err := cat.Get("urn:eudi:pid:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	tm, err := sdjwtvc.FromScheme(def)
	if err != nil {
		t.Fatalf("FromScheme: %v", err)
	}

	if tm.VCT != "urn:eudi:pid:1" {
		t.Errorf("vct = %q, want urn:eudi:pid:1", tm.VCT)
	}
	if len(tm.Claims) == 0 {
		t.Fatal("expected claims, got none")
	}

	var familyName *sdjwtvc.ClaimMeta
	for i := range tm.Claims {
		if len(tm.Claims[i].Path) == 1 && tm.Claims[i].Path[0] == "family_name" {
			familyName = &tm.Claims[i]
		}
	}
	if familyName == nil {
		t.Fatal("family_name claim not found in Type Metadata")
	}
	if !familyName.Mandatory {
		t.Error("family_name should be mandatory")
	}
	if familyName.SD != sdjwtvc.SDAlways {
		t.Errorf("family_name sd = %q, want always", familyName.SD)
	}
}

func TestFromSchemeMapsDisclosabilityToSpecVocabulary(t *testing.T) {
	cat := loadBundled(t)
	def, err := cat.Get("urn:eudi:pid:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	tm, err := sdjwtvc.FromScheme(def)
	if err != nil {
		t.Fatalf("FromScheme: %v", err)
	}

	var issuingAuthority *sdjwtvc.ClaimMeta
	for i := range tm.Claims {
		if len(tm.Claims[i].Path) == 1 && tm.Claims[i].Path[0] == "issuing_authority" {
			issuingAuthority = &tm.Claims[i]
		}
	}
	if issuingAuthority == nil {
		t.Fatal("issuing_authority claim not found")
	}
	// model.DisclosabilityMustNot ("MUST NOT") must map to the spec's "never".
	if issuingAuthority.SD != sdjwtvc.SDNever {
		t.Errorf("issuing_authority sd = %q, want never", issuingAuthority.SD)
	}
}

func TestFromSchemeFailsForMdocOnlyScheme(t *testing.T) {
	// A hypothetical mdoc-only definition (e.g. mDL, which has no SD-JWT VC
	// form): built by hand since none of our bundled definitions are
	// mdoc-only after the PID sd-jwt+mdoc merge (TS11 §4.3.1 models one
	// attestation type with multiple format Schemas, not one scheme id per
	// format).
	def := model.Definition{
		Rulebook: model.AttestationRulebook{AttestationType: "Hypothetical mdoc-only type"},
		Scheme: model.AttestationScheme{
			ID:               "org.example.mdoc-only.1",
			SupportedFormats: []model.CredentialFormat{model.FormatMDoc},
			Schemas: []model.FormatSchema{
				{Format: model.FormatMDoc, TypeIdentifier: "org.example.mdoc-only.1"},
			},
		},
	}

	if _, err := sdjwtvc.FromScheme(def); err == nil {
		t.Fatal("expected an error for a scheme with no dc+sd-jwt format")
	}
}
