// Package model defines the ARF 3.0 / ETSI TS 119 472 attestation data model:
// a human-readable AttestationRulebook paired with a machine-readable
// AttestationScheme, following the Attestation Rulebooks Catalog and TS11
// (Catalogue of Attributes and Catalogue of Attestations) conventions.
package model

// TrustAuthority resolves the trust framework applicable to an attestation
// scheme (ARF 3.0 / TS11 §4.3.3, OpenID4VP §6.1.1 trust mechanisms).
type TrustAuthority struct {
	FrameworkType TrustFrameworkType `json:"frameworkType"`
	Value         string             `json:"value"`
	IsLoTE        *bool              `json:"isLoTE,omitempty"`
}

// ClaimDefinition is the format-specific encoding of one attribute within an
// AttestationScheme, per Rulebook template §3 tables. Path addresses the
// claim (dot-notation segments for SD-JWT VC nesting, e.g. ["address",
// "locality"]; a single segment for mdoc namespace elements). Namespace and
// Disclosability are mutually exclusive: mdoc claims carry a namespace,
// SD-JWT VC claims carry a disclosability requirement (ARB_30).
type ClaimDefinition struct {
	DataIdentifier string         `json:"dataIdentifier"`
	Path           []string       `json:"path"`
	DataType       string         `json:"dataType"`
	Presence       Presence       `json:"presence"`
	Namespace      string         `json:"namespace,omitempty"`
	Disclosability Disclosability `json:"disclosability,omitempty"`
	// Enum restricts the claim's value to this fixed set, when the
	// Rulebook or an external spec (e.g. ETSI TS 119 472-1's fixed
	// "PID"/"QEAA"/... category-style values) mandates specific literal
	// values rather than a free-form DataType. Empty when unconstrained.
	Enum []string `json:"enum,omitempty"`
}

// FormatSchema is the ARF 3.0 / TS11 §4.3.2 Schema: the claim set for one
// issuance format of an attestation type. TypeIdentifier is the `vct` for
// FormatSDJWTVC or the mdoc doctype for FormatMDoc (TS11 §4.3.4).
type FormatSchema struct {
	Format         CredentialFormat  `json:"format"`
	TypeIdentifier string            `json:"typeIdentifier"`
	Claims         []ClaimDefinition `json:"claims"`
}

// Claim looks up a claim definition by its data identifier. Returns nil if
// not found.
func (s FormatSchema) Claim(dataIdentifier string) *ClaimDefinition {
	for i := range s.Claims {
		if s.Claims[i].DataIdentifier == dataIdentifier {
			return &s.Claims[i]
		}
	}
	return nil
}

// AttestationScheme is the ARF 3.0 / TS11 §4.3.1 SchemaMeta: the
// machine-readable attestation schema. Pairs with a human-readable
// AttestationRulebook via RulebookURI.
type AttestationScheme struct {
	// ID is this registry's lookup/URL identifier — a human-readable value
	// (a vct or mdoc doctype), used in GET /api/v1/schemes/{id}. TS11
	// §4.3.1 instead specifies SchemaMeta.id as an opaque UUID assigned by
	// the catalogue provider at registration time; CatalogueID carries
	// that. This registry deliberately keeps ID as the primary lookup key
	// since it has no external registration workflow (there is no separate
	// "the Commission assigns you a UUID" step here) and a human-readable
	// URL is far more usable for a small, internally-run registry — see
	// docs/compliance/*.md for the full rationale.
	ID                 string             `json:"id"`
	CatalogueID        string             `json:"catalogueId,omitempty"`
	Version            string             `json:"version"`
	RulebookURI        string             `json:"rulebookUri,omitempty"`
	TrustedAuthorities []TrustAuthority   `json:"trustedAuthorities,omitempty"`
	AttestationLoS     AssuranceLevel     `json:"attestationLoS"`
	BindingType        BindingType        `json:"bindingType"`
	SupportedFormats   []CredentialFormat `json:"supportedFormats"`
	Schemas            []FormatSchema     `json:"schemas"`
}

// SchemaFor returns the FormatSchema for the given format. Returns nil if the
// scheme does not support that format.
func (s AttestationScheme) SchemaFor(format CredentialFormat) *FormatSchema {
	for i := range s.Schemas {
		if s.Schemas[i].Format == format {
			return &s.Schemas[i]
		}
	}
	return nil
}

// AttestationRulebook is the human-readable Attestation Rulebook metadata
// (ARF 3.0 / TS11 §4.2, CIR for EAAs Art. 8). It mirrors the required
// elements of the Attestation Rulebook template without transcribing the
// full prose document — DocumentURI points at that.
//
// https://github.com/eu-digital-identity-wallet/eudi-doc-attestation-rulebooks-catalog/blob/main/template/attestation-rulebook-template.md
type AttestationRulebook struct {
	AttestationType string              `json:"attestationType"`
	Status          string              `json:"status"`
	Version         string              `json:"version"`
	Contact         string              `json:"contact,omitempty"`
	LegalBasis      []string            `json:"legalBasis,omitempty"`
	Category        AttestationCategory `json:"category,omitempty"`
	DocumentURI     string              `json:"documentUri,omitempty"`
	// Revocation is SHALL-required by the Rulebook template §6: every
	// Rulebook must state whether attestations are short-lived enough that
	// revocation is never necessary, or which revocation mechanism applies.
	Revocation RevocationMethod `json:"revocation,omitempty"`
	// RevocationListURL is the domain/URL at which Relying Parties can
	// retrieve the relevant Attestation Status List or Attestation
	// Revocation List, SHALL-required by §6 for revocable attestations.
	// Empty when Revocation is RevocationNotApplicableShortLived.
	RevocationListURL string `json:"revocationListUrl,omitempty"`
	// TrustAnchorDescription is the §5 (Trust anchors) prose: how a Relying
	// Party obtains and uses the trust anchor to verify this attestation
	// type. SHOULD-required for non-qualified EAAs (ARB_26); PIDs/QEAAs/
	// Pub-EAAs instead rely on the LoTL/Trusted List mechanism described in
	// ARF §6.6.3.6 and referenced via Scheme.TrustedAuthorities.
	TrustAnchorDescription string `json:"trustAnchorDescription,omitempty"`
}

// Definition is one catalogue entry: the pairing of a human-readable
// AttestationRulebook with its machine-readable AttestationScheme, as
// registered together in the Catalogue of Attestations (ARF 3.0 / TS11 §4).
type Definition struct {
	Rulebook AttestationRulebook `json:"rulebook"`
	Scheme   AttestationScheme   `json:"scheme"`
}
