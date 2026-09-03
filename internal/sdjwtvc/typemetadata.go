// Package sdjwtvc converts this registry's internal attestation model into
// SD-JWT VC Type Metadata Documents, as defined by the SD-JWT VC
// specification §"SD-JWT VC Type Metadata"
// (https://www.ietf.org/archive/id/draft-ietf-oauth-sd-jwt-vc-latest.html#name-sd-jwt-vc-type-metadata).
//
// This registry acts as the "Registry" retrieval method described in that
// section: a Consumer that does not resolve `vct` as an HTTPS URL can fetch
// the Type Metadata for a type from here instead, in exactly the format
// defined by the spec — not this registry's own internal JSON shape.
package sdjwtvc

import (
	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
)

// TypeMetadata is a Type Metadata Document per the SD-JWT VC "Type Metadata
// Format" section. Only the properties this registry can currently
// populate are included; the spec allows additional properties, and
// consumers must ignore ones they don't understand.
type TypeMetadata struct {
	VCT         string      `json:"vct"`
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Display     []Display   `json:"display,omitempty"`
	Claims      []ClaimMeta `json:"claims,omitempty"`
}

// Display is one locale's display information for the type, per
// "Display Metadata".
type Display struct {
	Locale      string `json:"locale"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ClaimMeta is one claim's metadata, per the "Claim Metadata" section.
type ClaimMeta struct {
	Path      []string            `json:"path"`
	Display   []ClaimDisplay      `json:"display,omitempty"`
	Mandatory bool                `json:"mandatory,omitempty"`
	SD        SelectiveDisclosure `json:"sd,omitempty"`
}

// ClaimDisplay is one locale's display information for a claim, per
// "Claim Display Metadata".
type ClaimDisplay struct {
	Locale      string `json:"locale"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// SelectiveDisclosure is the `sd` property's value, per "Claim Selective
// Disclosure Metadata". Distinct from this registry's internal
// model.Disclosability, whose MUST/MAY/MUST NOT values are Rulebook
// template wording, not the spec's own always/allowed/never vocabulary.
type SelectiveDisclosure string

const (
	SDAlways  SelectiveDisclosure = "always"
	SDAllowed SelectiveDisclosure = "allowed"
	SDNever   SelectiveDisclosure = "never"
)

// FromScheme builds a Type Metadata Document for the SD-JWT VC FormatSchema
// of an attestation definition. Returns an error if the scheme has no
// SD_JWT_VC format — Type Metadata is meaningless for mdoc.
func FromScheme(def model.Definition) (*TypeMetadata, error) {
	schema := def.Scheme.SchemaFor(model.FormatSDJWTVC)
	if schema == nil {
		return nil, errNoSDJWTFormat(def.Scheme.ID)
	}

	claims := make([]ClaimMeta, 0, len(schema.Claims))
	for _, c := range schema.Claims {
		claims = append(claims, ClaimMeta{
			Path:      c.Path,
			Mandatory: c.Presence == model.PresenceMandatory,
			SD:        toSelectiveDisclosure(c.Disclosability),
		})
	}

	return &TypeMetadata{
		VCT:         schema.TypeIdentifier,
		Name:        def.Rulebook.AttestationType,
		Description: def.Rulebook.AttestationType,
		Display: []Display{
			{Locale: "en-US", Name: def.Rulebook.AttestationType},
		},
		Claims: claims,
	}, nil
}

// toSelectiveDisclosure maps this registry's Rulebook-template-style
// Disclosability (MUST/MAY/MUST NOT, ARB_30 wording) onto the SD-JWT VC
// spec's own sd vocabulary (always/allowed/never). Empty/unset maps to
// "allowed", the spec's documented default.
func toSelectiveDisclosure(d model.Disclosability) SelectiveDisclosure {
	switch d {
	case model.DisclosabilityMust:
		return SDAlways
	case model.DisclosabilityMustNot:
		return SDNever
	default:
		return SDAllowed
	}
}

type errNoSDJWTFormat string

func (e errNoSDJWTFormat) Error() string {
	return "attestation scheme " + string(e) + " has no dc+sd-jwt format; Type Metadata is only defined for SD-JWT VC"
}
