package catalogue

import (
	"fmt"

	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
)

// Validate checks a claim value map against a FormatSchema: every mandatory
// claim must be present, and no claim outside the schema may be supplied.
// Claims are addressed by their top-level path segment; nested structure
// within a claim (e.g. address.locality) is not descended into. Returns the
// violations found; nil if claims satisfies the schema.
func Validate(schema model.FormatSchema, claims map[string]any) []string {
	var violations []string

	knownKeys := make(map[string]bool, len(schema.Claims))
	for _, claim := range schema.Claims {
		topLevelKey := claim.Path[0]
		knownKeys[topLevelKey] = true

		if claim.Presence == model.PresenceMandatory {
			if _, present := claims[topLevelKey]; !present {
				violations = append(violations, fmt.Sprintf("missing mandatory claim: %s", claim.DataIdentifier))
			}
		}
	}

	for suppliedKey := range claims {
		if !knownKeys[suppliedKey] {
			violations = append(violations, fmt.Sprintf("claim not defined in schema %s: %s", schema.TypeIdentifier, suppliedKey))
		}
	}

	return violations
}
