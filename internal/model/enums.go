package model

// CredentialFormat is the OID4VCI issuance format identifier.
type CredentialFormat string

const (
	FormatSDJWTVC CredentialFormat = "dc+sd-jwt"
	FormatMDoc    CredentialFormat = "mso_mdoc"
)

// AttestationCategory is the ETSI TS 119 472-1 `category` attribute value,
// indicating the legal nature of an attestation. PIDs use their own
// `attestation_legal_category` attribute (ARF Annex 3.01) instead of this one.
type AttestationCategory string

const (
	CategoryQEAA            AttestationCategory = "urn:etsi:esi:eaa:eu:qualified"
	CategoryPubEAA          AttestationCategory = "urn:etsi:esi:eaa:eu:pub"
	CategoryNonQualifiedEAA AttestationCategory = "eaa:eu:non-qualified"
)

// Presence states whether an attribute or claim is required, per the
// Attestation Rulebook template §2.
type Presence string

const (
	PresenceMandatory   Presence = "mandatory"
	PresenceOptional    Presence = "optional"
	PresenceConditional Presence = "conditional"
)

// Disclosability is the SD-JWT VC selective disclosure requirement for a
// claim (Rulebook template §3.2, ARB_30). Not applicable to mdoc claims.
type Disclosability string

const (
	DisclosabilityMust    Disclosability = "MUST"
	DisclosabilityMay     Disclosability = "MAY"
	DisclosabilityMustNot Disclosability = "MUST NOT"
)

// BindingType is the ARF 3.0 / TS11 SchemaMeta.bindingType: the cryptographic
// binding required for issuance.
type BindingType string

const (
	BindingClaim     BindingType = "claim"
	BindingKey       BindingType = "key"
	BindingBiometric BindingType = "biometric"
	BindingNone      BindingType = "none"
)

// AssuranceLevel is the ARF 3.0 / TS11 SchemaMeta.attestationLoS, per
// OpenID4VCI Annex D.2.
type AssuranceLevel string

const (
	AssuranceHigh          AssuranceLevel = "iso_18045_high"
	AssuranceModerate      AssuranceLevel = "iso_18045_moderate"
	AssuranceEnhancedBasic AssuranceLevel = "iso_18045_enhanced-basic"
	AssuranceBasic         AssuranceLevel = "iso_18045_basic"
)

// TrustFrameworkType is the applicable trust model for a TrustAuthority.
type TrustFrameworkType string

const (
	FrameworkAKI              TrustFrameworkType = "aki"
	FrameworkETSITL           TrustFrameworkType = "etsi_tl"
	FrameworkOpenIDFederation TrustFrameworkType = "openid_federation"
)
