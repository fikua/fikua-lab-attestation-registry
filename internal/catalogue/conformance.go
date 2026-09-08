package catalogue

import (
	"fmt"

	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
)

// ValidateConformance checks one AttestationDefinition against the ARF 3.0
// Annex 2 Topic 12 rules (FR-13) that this registry can verify on its own,
// without an external ecosystem-wide catalogue: enum membership for
// attestationLoS/bindingType/category, W3C VC restricted to non-qualified
// EAA (ARB_01a/ARB_04), and id uniqueness within the loaded catalogue
// (ARB_05 — ecosystem-wide uniqueness itself is out of reach without a real
// registration authority; see docs/compliance/*.md). Returns the violations
// found; nil if def is conformant.
//
// This is a pre-publish gate, not the claim-presence check Validate does —
// call it once per definition before adding it to the bundled catalogue
// (enforced today by TestAllBundledDefinitionsAreConformant), not per
// issued credential.
func ValidateConformance(def model.Definition, allIDs map[string]int) []string {
	var violations []string

	if !validAssuranceLevel[def.Scheme.AttestationLoS] {
		violations = append(violations, fmt.Sprintf("%s: attestationLoS %q is not a valid TS11 §4.3.1 value", def.Scheme.ID, def.Scheme.AttestationLoS))
	}
	if !validBindingType[def.Scheme.BindingType] {
		violations = append(violations, fmt.Sprintf("%s: bindingType %q is not a valid TS11 §4.3.1 value", def.Scheme.ID, def.Scheme.BindingType))
	}
	if def.Rulebook.Category != "" && !validCategory[def.Rulebook.Category] {
		violations = append(violations, fmt.Sprintf("%s: category %q is not a valid ARB_11/ARB_12 value", def.Scheme.ID, def.Rulebook.Category))
	}

	if def.Rulebook.Category != model.CategoryNonQualifiedEAA {
		for _, format := range def.Scheme.SupportedFormats {
			if isW3CFormat[format] {
				violations = append(violations, fmt.Sprintf("%s: format %q (W3C VC) is only admissible for a non-qualified EAA (ARB_01a/ARB_04), category is %q", def.Scheme.ID, format, def.Rulebook.Category))
			}
		}
	}

	if allIDs[def.Scheme.ID] > 1 {
		violations = append(violations, fmt.Sprintf("%s: scheme id is registered more than once in this catalogue (ARB_05)", def.Scheme.ID))
	}

	if def.Rulebook.Revocation == model.RevocationNotApplicableShortLived {
		if def.Rulebook.MaxValiditySeconds <= 0 {
			violations = append(violations, fmt.Sprintf("%s: revocation is not_applicable_short_lived but maxValiditySeconds is unset — VCR_01b requires a declared validity period to justify skipping revocation", def.Scheme.ID))
		} else if def.Rulebook.MaxValiditySeconds > maxShortLivedSeconds {
			violations = append(violations, fmt.Sprintf("%s: maxValiditySeconds %d exceeds the 24h short-lived ceiling VCR_01b requires for not_applicable_short_lived (ETSI EN 319 411-1 REV-6.2.4-03A)", def.Scheme.ID, def.Rulebook.MaxValiditySeconds))
		}
	}

	return violations
}

// maxShortLivedSeconds is the 24-hour ceiling ARF 3.0 Topic 7 VCR_01b sets
// for an attestation to justify "short-lived, so revocation is never
// necessary" instead of implementing a real revocation mechanism.
const maxShortLivedSeconds = 24 * 60 * 60

var validAssuranceLevel = map[model.AssuranceLevel]bool{
	model.AssuranceHigh:          true,
	model.AssuranceModerate:      true,
	model.AssuranceEnhancedBasic: true,
	model.AssuranceBasic:         true,
}

var validBindingType = map[model.BindingType]bool{
	model.BindingClaim:     true,
	model.BindingKey:       true,
	model.BindingBiometric: true,
	model.BindingNone:      true,
}

var validCategory = map[model.AttestationCategory]bool{
	model.CategoryQEAA:            true,
	model.CategoryPubEAA:          true,
	model.CategoryNonQualifiedEAA: true,
}

// isW3CFormat marks the OpenID4VCI W3C VC format identifiers (TS11 §4.3.1
// supportedFormats); dc+sd-jwt and mso_mdoc are not W3C formats.
var isW3CFormat = map[model.CredentialFormat]bool{
	model.CredentialFormat("jwt_vc_json"):    true,
	model.CredentialFormat("jwt_vc_json-ld"): true,
	model.CredentialFormat("ldp_vc"):         true,
}

// IDCounts tallies how many times each scheme id appears across defs, for
// use as ValidateConformance's allIDs argument.
func IDCounts(defs []model.Definition) map[string]int {
	counts := make(map[string]int, len(defs))
	for _, d := range defs {
		counts[d.Scheme.ID]++
	}
	return counts
}
