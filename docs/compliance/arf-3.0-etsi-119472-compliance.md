# ARF 3.0 / ETSI TS 119 472 / CIR 2024/2977 Compliance Review

**Last reviewed:** 2026-09-03, against commit `8922715` and the fixes described below.

This document tracks how the bundled attestation definitions
(`data/attestations/*.json`) and the domain model (`internal/model`) map to
the normative sources this registry claims to implement:

- [ARF 3.0](https://github.com/eu-digital-identity-wallet/eudi-doc-architecture-and-reference-framework)
- [PID Rulebook, Annex 3.01](https://github.com/eu-digital-identity-wallet/eudi-doc-attestation-rulebooks-catalog/blob/main/rulebooks/pid/pid-rulebook.md) (v1.7)
- [Attestation Rulebook template](https://github.com/eu-digital-identity-wallet/eudi-doc-attestation-rulebooks-catalog/blob/main/template/attestation-rulebook-template.md) (v1.5)
- [TS11 — Catalogue of Attributes and Catalogue of Attestations](https://github.com/eu-digital-identity-wallet/eudi-doc-standards-and-technical-specifications/blob/main/docs/technical-specifications/ts11-interfaces-and-formats-for-catalogue-of-attributes-and-catalogue-of-schemes.md) (v1.0.1)
- ETSI TS 119 472-1 (attribute `category`)
- [SD-JWT VC — Type Metadata](https://www.ietf.org/archive/id/draft-ietf-oauth-sd-jwt-vc-latest.html#name-sd-jwt-vc-type-metadata)
- CIR 2024/2977 (PID attribute set)

## Summary

| Area | Status |
|---|---|
| PID modeled as one attestation type with two format schemas (TS11 §4.3.1) | ✅ Fixed — see [Modeling fix: PID as one scheme, two formats](#modeling-fix-pid-as-one-scheme-two-formats) |
| PID mandatory/optional attributes match CIR 2024/2977 / Rulebook §2 | ✅ |
| SD-JWT VC claim names match Rulebook §4.1 (IANA/OIDC/EKYC names) | ✅ |
| mdoc attribute identifiers/namespace match Rulebook §3.1 | ✅ |
| `category` present on the EAA, absent on the PID (PID uses `attestation_legal_category` instead) | ✅ |
| `attestation_legal_category` constrained to `"PID"` | ✅ Fixed |
| `sex` constrained to ISO/IEC 5218 values | ✅ Fixed |
| Revocation declared per Rulebook template §6 (SHALL) | ✅ Fixed |
| Trust anchor description per Rulebook template §5 (SHOULD for non-qualified EAA) | ✅ Fixed |
| `SchemaMeta.id` as an opaque catalogue UUID (TS11 §4.3.1) | ⚠️ Deviation, documented — see [Known deviation: catalogue id](#known-deviation-catalogue-id-vs-lookup-id) |
| Real SD-JWT VC Type Metadata Document, not just our internal claim shape | ✅ (`internal/sdjwtvc`, `/api/v1/schemes/{id}/type-metadata`) |
| Real ISO 23220-2 mdoc DocType document | ❌ Not implemented — see [Open gap: mdoc DocType](#open-gap-mdoc-doctype-document) |
| `tstr` length/type constraints (CDDL) enforced, not just documented | ❌ Not implemented — see [Open gap: CDDL-level constraints](#open-gap-cddl-level-constraints) |
| JWT protocol claims (`iss`, `cnf`, `exp`, `nbf`) represented | ❌ Not modeled — see [Open gap: JWT protocol claims](#open-gap-jwt-protocol-claims) |
| Barcelona padró: EAA-specific requirements (category, disclosability, trust anchor, revocation) | ✅ |
| ETSI TS 119 472-2 mdoc namespace for `category` (`org.etsi.01947201.010101`) | 📝 Documented as a constant (`model.ETSICategoryMdocNamespace`), unused — no bundled EAA supports mdoc yet |

## Modeling fix: PID as one scheme, two formats

**Before:** `eudi-pid-sdjwt.json` and `eudi-pid-mdoc.json` were two separate
catalogue entries with two different `scheme.id`s (`urn:eudi:pid:1` and
`eu.europa.ec.eudi.pid.1`).

**Why this was wrong:** TS11 §4.3.1 models a `SchemaMeta` (→ our
`AttestationScheme`) per *attestation type*, with `supportedFormats` and an
array of `Schema` objects (→ our `FormatSchema`) — one per format. The PID
Rulebook itself treats SD-JWT VC and mdoc encodings as chapters 3 and 4 of
a *single* document, not two attestation types. The `urn:eudi:pid:1` vct
and `eu.europa.ec.eudi.pid.1` mdoc doctype are two identifiers for the
*same* logical attestation.

**Fix:** merged into `data/attestations/eudi-pid.json`, one
`AttestationDefinition` with `scheme.schemas` containing both a `dc+sd-jwt`
and a `mso_mdoc` `FormatSchema`, each with its own `typeIdentifier`. The
mdoc doctype is no longer independently reachable via
`GET /api/v1/schemes/{id}` — see `TestPidMdocTypeIdentifierIsNotASeparateSchemeID`.

## Known deviation: catalogue id vs. lookup id

TS11 §4.3.1 specifies `SchemaMeta.id` as *"unique identifier (UUID),
provided by the server of the catalogue provider"* — an opaque value
assigned at registration time, not a human-readable identifier.

This registry's `AttestationScheme.ID` is instead the human-readable vct
or mdoc doctype (`urn:eudi:pid:1`), used directly as the
`GET /api/v1/schemes/{id}` lookup key and in the UI's URLs.

**Why:** this registry has no external registration workflow — there is no
"submit your attestation type to the Commission and receive a UUID back"
step, since it's a small, internally-run catalogue for `fikua-lab-issuer`
and `fikua-lab-verifier`. A human-readable URL (`/rulebooks/urn:eudi:pid:1`)
is far more usable for browsing and debugging than a UUID would be, and
nothing external depends on this registry's ids being UUIDs.

**Mitigation:** added `AttestationScheme.CatalogueID` as a separate field
carrying a real, fixed UUID per definition, so the TS11-shaped field
exists and is populated (`TestEveryDefinitionHasACatalogueUUID`) — callers
that need strict TS11 compliance can read `catalogueId`; this registry's
own API and UI continue to use the human-readable `id`.

## Open gap: mdoc DocType document

TS11 §4.3.4 requires each format's schema document to be format-native:
Type Metadata for `dc+sd-jwt` (now implemented, see below), an
**ISO/IEC 23220-2 DocType document** for `mso_mdoc`. No bundled definition
currently needs this (the only mdoc format we define, PID, is a
well-known EU-wide doctype), but if a domestic or lab-specific mdoc type
is ever added, a `/api/v1/schemes/{id}/doctype` endpoint analogous to
`internal/sdjwtvc.FromScheme` would be needed. Not built yet — no
consumer needs it today.

## Open gap: CDDL-level constraints

The PID Rulebook mandates concrete encoding constraints for mdoc
attributes (Rulebook §3.1.2): `tstr` values SHALL have a maximum length of
150 characters, `full-date`/`tdate` follow specific RFC 3339 profiles,
canonical CBOR rules apply, etc. `ClaimDefinition.DataType` is a free-text
description (e.g. `"tstr"`, `"full-date"`) for documentation purposes —
this registry does not (and currently has no consumer that would) validate
issued credential values against these constraints. `SchemeValidator`
(`internal/catalogue/validator.go`) only checks claim *presence*
(mandatory/unknown), not value-level type or length constraints.

## Open gap: JWT protocol claims

An SD-JWT VC credential requires standard JWT claims outside this
registry's domain model: `iss`, `cnf` (key binding), `exp`/`nbf`
(technical validity — distinct from the PID Rulebook's *administrative*
`date_of_expiry`/`date_of_issuance`, which are real domain claims and
are modeled), `vct#integrity`, etc. `FormatSchema.Claims` only enumerates
domain-specific claims (the ones an issuer/verifier author cares about),
not the protocol envelope. This is a deliberate scope boundary — protocol
claims are `fikua-lab-issuer`'s responsibility to add when it builds the
actual credential, not something a claim *catalogue* needs to enumerate —
but it means this registry's schema is not, by itself, a complete
description of a valid SD-JWT VC token.

## Verified compliant details

- **PID mandatory attributes** (`family_name`, `given_name`, `birth_date`,
  `birth_place`, `nationality`, `portrait`) present in both format schemas,
  matching CIR 2024/2977 / Rulebook §2.2.
- **PID optional attributes** (resident address fields, birth names, sex,
  email, phone, personal administrative number, expiry/issuance dates,
  document number, jurisdiction, trust anchor, legal category) match
  Rulebook §2.3–§2.6.
- **SD-JWT VC claim names** match Rulebook §4.1 exactly: `birthdate` (not
  `birth_date`), `place_of_birth`, `nationalities`, `picture`,
  `birth_family_name`/`birth_given_name`, `address.*` sub-claims,
  `date_of_expiry`/`date_of_issuance` as Private Names.
  `attestation_legal_category` and `sex` are now enum-constrained
  (`"PID"` and ISO/IEC 5218's `0,1,2,3,4,5,6,9` respectively).
- **mdoc attribute identifiers and namespace** match Rulebook §3.1.2: both
  doctype and namespace are `eu.europa.ec.eudi.pid.1`, attribute
  identifiers are identical to the data identifiers (no renaming).
- **`category` (ETSI TS 119 472-1)**: correctly absent from the PID
  (which uses its own `attestation_legal_category` per Rulebook §2.6),
  correctly present on the Barcelona padró EAA with a valid enum value
  (`eaa:eu:non-qualified`) and `disclosability: MUST NOT`, matching the
  Rulebook template's worked example in §3.2.
- **Revocation** (Rulebook template §6, SHALL): PID declares
  `attestation_status_list` with a list URL; the padró (a lab prototype
  with no revocation infrastructure) declares
  `not_applicable_short_lived`. Enforced by
  `TestEveryDefinitionDeclaresRevocation`.
- **Trust anchors** (Rulebook template §5): PID relies on the LoTL
  mechanism (already modeled via `Scheme.TrustedAuthorities`); the padró
  documents, in prose, what a production issuer would need to do (ARB_26)
  since it has no real trust anchor infrastructure.
- **`cryptographically_bound_to`** (ARB_28, Rulebook template §4): present
  on the padró, correctly `mandatory`/`MUST NOT` disclosable, since a
  municipal registration attestation presupposes a verified PID.

## Deliberately out of scope

- **Qualified/non-qualified EAA governance** (registration with the
  Commission's real Catalogue of Attestations, `SchemaMeta` UUID
  assignment, ETSI TS 119 478 authentic-source verification) — this is a
  lab/demo registry, not a production catalogue provider.
- **CIR 2024/2977 legal interpretation** beyond what the PID Rulebook
  already encodes — the Rulebook is treated as the authoritative
  encoding-independent source; this registry does not re-derive
  requirements from the raw CIR text.
