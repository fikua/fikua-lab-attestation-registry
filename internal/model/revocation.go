package model

// RevocationMethod is how a Relying Party can check whether an attestation
// has been revoked, per Attestation Rulebook template §6 (Revocation). The
// template requires every Rulebook to state one of these explicitly: either
// attestations are short-lived enough (<=24h) that revocation is never
// needed, or one of the two listed mechanisms applies.
type RevocationMethod string

const (
	// RevocationNotApplicableShortLived: validity period <=24h, so
	// revocation checking is never necessary.
	RevocationNotApplicableShortLived RevocationMethod = "not_applicable_short_lived"
	RevocationStatusList              RevocationMethod = "attestation_status_list"
	RevocationRevocationList          RevocationMethod = "attestation_revocation_list"
)
