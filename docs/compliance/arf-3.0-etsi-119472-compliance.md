# ARF 3.0 / ETSI TS 119 472 / CIR 2024/2977 Compliance Review

**Last reviewed:** 2026-09-08, against a full-text TS11 audit (§1–§6, not just
§4.3), a cross-check against Altia's EUD-30 SRS (a sibling ARF/TS11
implementation), a full ARF 3.0 Annex 2 Topic 12 (ARB_xx) / eIDAS 2.0
(Regulation (EU) 2024/1183) / CIR (EU) 2025/1569 review of the Barcelona
padró specifically (ahead of an institutional offer to the Ajuntament), and
the fixes described below.

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
| Real SD-JWT VC Type Metadata Document, not just our internal claim shape | ✅ (`internal/sdjwtvc`, `/api/v1/schemes/{id}/{version}/type-metadata`) |
| Real ISO 23220-2 mdoc DocType document | ❌ Not implemented — see [Open gap: mdoc DocType](#open-gap-mdoc-doctype-document) |
| `tstr` length/type constraints (CDDL) enforced, not just documented | ❌ Not implemented — see [Open gap: CDDL-level constraints](#open-gap-cddl-level-constraints) |
| JWT protocol claims (`iss`, `cnf`, `exp`, `nbf`) represented | ❌ Not modeled — see [Open gap: JWT protocol claims](#open-gap-jwt-protocol-claims) |
| Barcelona padró: EAA-specific requirements (category, disclosability, trust anchor, revocation) | ✅ |
| ETSI TS 119 472-2 mdoc namespace for `category` (`org.etsi.01947201.010101`) | 📝 Documented as a constant (`model.ETSICategoryMdocNamespace`), unused — no bundled EAA supports mdoc yet |
| `SchemaMeta.schemaURIs` as `{formatIdentifier, uri}` pointers, not inline claims (TS11 §4.3.2) | ✅ Fixed — see [Modeling fix: schemaURIs as pointers, not containers](#modeling-fix-schemauris-as-pointers-not-containers) |
| Barcelona padró: `rulebookUri` present (TS11 §4.3.1, `[1..1]` required) | ✅ Fixed — was entirely absent from `padro-barcelona.json` |
| Catalogue of Attributes (TS11 §2/§3, OOTS/SDG-hosted EC infrastructure) | ⚠️ Deviation, documented — see [Deliberately out of scope](#deliberately-out-of-scope) |
| `GET /schemas` query-parameter filtering, pagination, JWS-signed response (TS11 §5) | ❌ Not implemented — see [Open gap: API §5 read-surface features](#open-gap-api-5-read-surface-features) |
| `supportedFormats`/`formatIdentifier` only cover 2 of 5 TS11 values (`jwt_vc_json`, `jwt_vc_json-ld`, `ldp_vc` absent) | ⚠️ Deviation, documented — see [Deliberately out of scope](#deliberately-out-of-scope) |
| Pre-publish conformance gate: `attestationLoS`/`bindingType`/`category` enum membership, W3C VC restricted to non-qualified EAA (ARB_01a/ARB_04), id uniqueness within the loaded catalogue (ARB_05) | ✅ Fixed — see [Modeling fix: pre-publish conformance gate](#modeling-fix-pre-publish-conformance-gate) |
| `schemaURIs[].uri` carries a `#integrity` (SRI) suffix (SD-JWT VC §7, TS11 §4.3.1/§4.3.2) | ✅ Fixed — see [Modeling fix: integrity suffix on schemaURIs](#modeling-fix-integrity-suffix-on-schemauris) |
| Published format documents are version-addressed and immutable (FR-18/19/20-style, no normative TS11 requirement but matches Altia's EUD-30 SRS) | ✅ Fixed — see [Modeling fix: version-addressed type-metadata](#modeling-fix-version-addressed-type-metadata) |
| Multi-tenant authorship isolation (Altia EUD-30 NFR-S-01/FR-05) | ⚠️ Deliberately out of scope — see [Deliberately out of scope](#deliberately-out-of-scope) |
| CDN / edge caching, hash-addressed immutable content (Altia EUD-30 NFR-A-01) | ⚠️ Deliberately out of scope — see [Deliberately out of scope](#deliberately-out-of-scope) |
| Padró: `category` claim renamed to `attestation_legal_category` (ARB_25 exact name, SHALL) | ✅ Fixed |
| Padró: `maxValiditySeconds` backing `not_applicable_short_lived` (ARF Topic 7 VCR_01b, ≤24h) | ✅ Fixed, gated by `ValidateConformance` |
| Padró: proximity use case analysis documented (ARB_02, SHALL the analysis happen, not just the conclusion) | ✅ Fixed |
| Padró: personal data separation statement (eIDAS 2.0 Art. 45h, SHALL for non-qualified EAA too) | ✅ Fixed |
| Padró: `eaa:eu:non-qualified` vs. `eaa:eu:pub` classification | ✅ Deliberate, investigated in full — see [Deliberate choice: non-qualified EAA, not Pub-EAA](#deliberate-choice-non-qualified-eaa-not-pub-eaa-2026-09-08) |

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

## Modeling fix: schemaURIs as pointers, not containers

**Before:** `internal/model.FormatSchema` — documented in its own comment as
*"the ARF 3.0 / TS11 §4.3.2 Schema"* — embedded the full claim list
(`Claims []ClaimDefinition`) inline, and had no `uri` field at all.

**Why this was wrong:** TS11 §4.3.2 `Schema` is a two-field pointer:
`formatIdentifier` and `uri` only — *"persistent schema URI assigned for
each registered format"*. §4.3.4 confirms the claim content lives in a
separate, format-native document (SD-JWT VC Type Metadata for `dc+sd-jwt`,
an ISO 23220-2 DocType for `mso_mdoc`) served *at* that `uri`, *"provided as
appendixes of the Attestation Rulebook."* §4.4's cross-reference table
independently corroborates this: attribute namespaces and definitions are
listed as living at `SchemaMeta.Schema.uri`, not inline in the index.
`FormatSchema` was conflating TS11's index layer (`SchemaMeta`/`Schema`)
with its content layer (the format-native documents `Schema.uri` points
at).

**Fix:** `internal/model.FormatSchema`'s doc comment now states plainly
that it is this registry's own internal representation, not the TS11
`Schema` sub-class — `Claims` stays (it's the source data
`internal/sdjwtvc.FromScheme` renders into a real Type Metadata Document,
and `fikua-lab-issuer`/`fikua-lab-verifier` consume it directly; removing
it would be a breaking API change with no compliance benefit). What
changed is the HTTP layer: `internal/httpapi` now projects
`AttestationScheme.Schemas` into a genuine, spec-shaped `schemaURIs` array
(`{formatIdentifier, uri}`) on every `GET /api/v1/schemes` and
`GET /api/v1/schemes/{id}` response, with `uri` computed from the request's
host/basePath (the `model` package does no I/O, so it can't know its own
deployed address) — `dc+sd-jwt` points at the existing
`/api/v1/schemes/{id}/{version}/type-metadata` endpoint; `mso_mdoc` gets an
empty `uri` until an ISO 23220-2 DocType generator exists (see
[Open gap: mdoc DocType document](#open-gap-mdoc-doctype-document)), rather
than a link to a document that doesn't exist. See
`TestGetSchemeIncludesSchemaURIs`.

## Modeling fix: pre-publish conformance gate

**Before:** `internal/catalogue/validator.go`'s `Validate` only checked
claim presence against a `FormatSchema` (mandatory claims present, no
unknown claims) — it says nothing about whether the scheme itself is
well-formed against ARF 3.0 Annex 2 Topic 12.

**Why this was a gap:** Altia's EUD-30 SRS (a sibling ARF/TS11
implementation, cross-checked 2026-09-08) specifies an explicit pre-publish
gate (F-03/FR-13/FR-14) that rejects a type violating: enum membership for
`category`/`attestationLoS`/`bindingType`, W3C VC restricted to
non-qualified EAA (`ARB_01a`/`ARB_04`), and scheme id uniqueness
(`ARB_05`). This registry had none of that — a malformed enum value or a
duplicate id would silently load.

**Fix:** `internal/catalogue/conformance.go` adds `ValidateConformance`,
checked against every bundled definition by
`TestAllBundledDefinitionsAreConformant` (`internal/catalogue/conformance_test.go`).
This registry has no separate "submit for publication" step — adding a
JSON file under `data/attestations/` and passing this test *is*
publishing — so the test suite is where the gate lives, not a runtime
check. `ARB_05` uniqueness is checked only within this registry's own
loaded catalogue, not the whole EUDI ecosystem: verifying that requires a
real registration authority this lab registry doesn't have (see
[Deliberately out of scope](#deliberately-out-of-scope)).

## Modeling fix: integrity suffix on schemaURIs

**Before:** `schemaURIs[].uri` was a plain URL with no way for a consumer
to verify the fetched document without trusting the transport.

**Why this was a gap:** TS11 §4.3.1/§4.3.2 state, verbatim, that
`rulebookURI`/`Schema.uri` *"MAY be suffixed with #integrity, the value of
which SHALL be an 'integrity metadata' string as defined in Section 3 of
[W3C SRI]"* — i.e. the fragment carries a `integrity=` key
(`#integrity=sha256-<base64>`), not just the bare hash. Altia's EUD-30 SRS
makes this a MUST for SD-JWT VC/W3C (FR-09, NFR-S-02), and its own
`attestation-scheme.json` spike (cross-checked 2026-09-08, e.g.
`.../sd-jwt.json#integrity=sha256-H3sCpB5Xa55MwjarfOtSPAE+HfMoHyl06N4/gtM9Sk0=`)
uses exactly this form — confirming the reading against a second
independent source, not just the raw spec text.

**Note — this fix has a prior, incorrect iteration:** the first pass at
this (2026-09-08, same day) appended a bare `#sha256-...` fragment with no
`integrity=` key, conflating this URL-fragment mechanism with SD-JWT VC
§7's *different* integrity mechanism (`vct#integrity`, a sibling JSON claim
*inside the token itself*, not a URL fragment at all — that one is
`fikua-lab-issuer`'s responsibility, not this registry's). Caught by
diffing against Altia's spike files before it shipped past this repo.

**Fix:** `internal/sdjwtvc.Integrity` computes a `sha256-<base64>` SRI
digest over the exact JSON bytes `writeJSON` serializes for the Type
Metadata Document. `internal/httpapi.formatDocumentURI` appends it to the
`dc+sd-jwt` `schemaURIs` entry as a `#integrity=sha256-...` suffix.
`mso_mdoc` has no integrity value — it has no generated document yet at
all (see
[Open gap: mdoc DocType document](#open-gap-mdoc-doctype-document));
mdoc/CBOR also has no standard `#integrity`-equivalent mechanism, which is
why NFR-S-02 in Altia's own SRS scopes the integrity guarantee to
SD-JWT VC/W3C only, not mdoc.

## Modeling fix: version-addressed type-metadata

**Before:** `GET /api/v1/schemes/{id}/type-metadata` always served
whatever the *current* bundled JSON said — publishing a new version of a
definition silently changed what that URL returned, with no way for an
already-issued credential's `vct` resolution to pin an exact version.

**Why this was a gap:** Altia's EUD-30 SRS requires (FR-18/19/20) that a
published version's artifact is address-by-version and never overwritten,
so that publishing a new version never invalidates credentials already
issued against an older one. TS11 itself doesn't mandate this explicitly,
but it's a reasonable property for any registry serving documents that
credentials point at long after issuance, and it costs little to have.

**Fix:** the route is now
`GET /api/v1/schemes/{id}/{version}/type-metadata` — `version` must match
`Scheme.Version` exactly, checked in `getTypeMetadata`
(`internal/httpapi/handlers.go`), 404 otherwise. There is deliberately no
unversioned alias: `schemaURIs[].uri` always names an exact version, and a
caller that wants "whatever the current version is" reads
`scheme.version` from `GET /api/v1/schemes/{id}` and builds the URL
itself. **This is a partial fix, not full FR-18 compliance**: this
registry does not yet keep *prior* versions' content addressable — bumping
`version` in a `data/attestations/*.json` file changes what that new
version's URL serves, but the *old* version's URL 404s rather than
continuing to serve the old content, since there is no version store, only
the current bundled JSON. Full immutability would need one (out of scope
for now — no bundled definition has been re-versioned in production yet,
so there's no real prior version to preserve). Checked no consumer needed
updating: `fikua-lab-issuer` (`internal/registryclient`) only calls
`GET /api/v1/schemes` and `GET /api/v1/schemes/{id}`, never
`/type-metadata` — that endpoint is resolved directly by wallets/Verifiers
from a credential's `vct`, not proxied through the issuer.

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

## Open gap: API §5 read-surface features

TS11 §5 specifies more for the read API than this registry currently
provides:

- **Query-parameter filtering** on `GET /schemas` — TS11 lists filterable
  fields (`id`, `supportedformat`, `attestationlos`, `bindingtype`,
  `trustedAuthoritiesFrameworkType`, `trustedAuthoritiesValue`,
  `schemauri`, `rulebookuri`) as query parameters. `listSchemes`
  (`internal/httpapi/handlers.go`) returns every definition unconditionally
  — no filtering.
- **Pagination** — TS11's `GET /schemas` response is documented as
  paginated; ours returns the full list in one response.
- **JWS-signed responses** — TS11 §5.3.1 specifies the method response body
  is JWS-signed. Ours is plain JSON.

**Why not fixed now:** none of `fikua-lab-issuer`/`fikua-lab-verifier`
(this registry's only consumers) need filtering, pagination, or signed
responses at the current catalogue size (2 definitions). Revisit if this
registry ever needs to serve as a drop-in for the literal TS11 Annex A.3
OpenAPI contract, or if the catalogue grows large enough that an
unconditional full dump becomes a real cost.

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
- **`attestation_legal_category` (ETSI TS 119 472-1's `category` attribute,
  ARB_25's exact required claim name)**: correctly absent from the PID
  (which instead uses its own `attestation_legal_category` semantics per
  Rulebook §2.6, a different attribute despite the shared name — see PID
  Rulebook v1.7 §2.6), correctly present on the Barcelona padró EAA with a
  valid enum value (`eaa:eu:non-qualified`) and `disclosability: MUST NOT`,
  matching the Rulebook template's worked example in §3.2. Renamed from a
  bare `category` claim to `attestation_legal_category` (2026-09-08) —
  ARB_25 requires that exact name (SHALL), not just equivalent semantics.
- **Revocation** (Rulebook template §6, SHALL; ARF 3.0 Topic 7 VCR_02 for
  non-qualified EAAs): PID declares `attestation_status_list` with a list
  URL; the padró declares `not_applicable_short_lived`, now with
  `maxValiditySeconds: 86400` backing that claim — VCR_01b's "short-lived,
  so revocation is never necessary" basis is only valid for a token
  validity of 24 hours or less (ETSI EN 319 411-1 REV-6.2.4-03A); an
  unbacked `not_applicable_short_lived` was an unverifiable assertion
  before this field existed. Enforced by
  `TestEveryDefinitionDeclaresRevocation` and
  `TestAllBundledDefinitionsAreConformant` (the latter now rejects
  `not_applicable_short_lived` with no `maxValiditySeconds`, or one
  exceeding 24h).
- **Trust anchors** (Rulebook template §5, ARB_26 SHOULD for non-qualified
  EAA): PID relies on the LoTL mechanism (already modeled via
  `Scheme.TrustedAuthorities`); the padró documents, in prose, what a
  production issuer would need to do (ARB_21/ARB_26 — publish a List of
  Trusted Entities per ETSI TS 119 602) since it has no real trust anchor
  infrastructure.
- **`cryptographically_bound_to`** (ARB_28, Rulebook template §4): present
  on the padró, correctly `mandatory`/`MUST NOT` disclosable, since a
  municipal registration attestation presupposes a verified PID.
- **Proximity use case analysis** (ARB_02, SHALL — the analysis must be
  performed and documented, not just have a correct conclusion): the padró
  now states, in `rulebook.proximityUseCaseAnalysis`, that no identified
  use case requires offline/proximity presentation today, and that
  `mso_mdoc` would need to be added if one is identified — previously
  `dc+sd-jwt`-only was correct in outcome but undocumented as a reasoned
  decision.
- **Personal data separation** (eIDAS 2.0 Art. 45h, SHALL — applies to
  non-qualified EAA providers too, not only qualified ones): the padró now
  states, in `rulebook.dataSeparationStatement`, that personal data used
  to issue it is not combined with other services' data.

## Deliberate choice: non-qualified EAA, not Pub-EAA (2026-09-08)

The Ajuntament de Barcelona, as keeper of the Padró municipal, is by
Spanish law (Ley 7/1985 art. 15-17) the public sector body responsible for
an authentic source — which makes it a *plausible* candidate for
`eaa:eu:pub` (electronic attestation of attributes issued by or on behalf
of a public sector body responsible for an authentic source, eIDAS 2.0
Art. 45f) rather than `eaa:eu:non-qualified`. This was investigated in
full (2026-09-08) before deciding to keep `non-qualified` — recorded here
so the choice isn't re-litigated from scratch, or silently drifted from,
later.

**Why `eaa:eu:pub` was not adopted, with exact requirements checked:**

- **Annex VII / Art. 45f(1)(b)** (Regulation (EU) 2024/1183) requires the
  issuer's qualified electronic signature or seal to come from a
  **qualified certificate for electronic seal (QSealC) carrying specific
  certified attributes**: that the issuing body is established under
  Union/national law as responsible for the authentic source (or
  designated to act on its behalf), data unambiguously representing that
  authentic source, and the specific law under which it acts. This is not
  an ordinary sede electrónica certificate — it is a new, Art. 45f-specific
  certificate type. The Ajuntament does not have one today.
- **Art. 45f(3) / CIR (EU) 2025/1569 Art. 5-6**: the public sector body
  must be *notified* by its Member State (Spain) to the Commission, with a
  conformity assessment report, before the Commission publishes it on the
  public list of Pub-EAA-issuing bodies. This notification/publication
  duty (CIR Art. 6) has applied since **2026-08-19** (CIR Art. 11) — it is
  already in force, not a future date — and the Ajuntament has not been
  through it. Declaring `eaa:eu:pub` without it would assert a status this
  registry cannot substantiate if checked against the Commission's
  published list.
- **Revocation stops being optional**: VCR_01/VCR_01b (SHALL for
  QEAA/Pub-EAA, unconditionally — unlike VCR_02's conditional SHALL for
  non-qualified EAA) and CIR Art. 4(3) require a real Attestation Status
  List or Attestation Revocation List for any Pub-EAA valid for more than
  24 hours. A certificat d'empadronament's realistic validity (months, not
  hours) rules out the `not_applicable_short_lived` basis this registry
  currently uses — adopting `pub` would require standing up real
  revocation infrastructure as a hard prerequisite, not an optional
  hardening step.
- **`trustedAuthorities` would need to point at the Pub-EAA Provider LoTE**
  (List of Trusted Entities per ETSI TS 119 602 + Annex H) rather than the
  generic, SHOULD-level trust anchor description this registry uses today
  for non-qualified EAAs (ARB_21/ARB_26) — another piece of infrastructure
  that doesn't exist yet.

**Conclusion:** none of these are modeling gaps this registry can close by
itself — they require the Ajuntament to obtain a Art. 45f-specific QSealC
and go through Spain's formal notification to the Commission, plus this
registry (or `fikua-lab-issuer`) standing up real revocation. Declaring
`eaa:eu:pub` today, without those, would be a legal-category claim this
setup cannot back up — worse for an institutional offer than the accurate
`eaa:eu:non-qualified`, whose legitimacy correctly rests on the cited
Spanish legal basis (Ley 7/1985, RD 1690/1986, Resolución INE/DGCG 2020)
plus ARF 3.0 Topic 12's generically-applicable requirements, not on a
CIR 2025/1569 obligation it doesn't actually have (CIR Art. 1 + Recital 14
confirm CIR 2025/1569 is voluntary, not binding, for non-qualified EAAs).
**Revisit this decision, deliberately, if/when the Ajuntament obtains the
Art. 45f QSealC and Spain completes the CIR Art. 5 notification** — at
that point `eaa:eu:pub` becomes both accurate and advantageous (SHALL
trust anchor instead of SHOULD, Commission-published verifiability).

## Deliberately out of scope

- **Qualified/non-qualified EAA governance** (registration with the
  Commission's real Catalogue of Attestations, `SchemaMeta` UUID
  assignment, ETSI TS 119 478 authentic-source verification) — this is a
  lab/demo registry, not a production catalogue provider.
- **CIR 2024/2977 legal interpretation** beyond what the PID Rulebook
  already encodes — the Rulebook is treated as the authoritative
  encoding-independent source; this registry does not re-derive
  requirements from the raw CIR text.
- **TS11 §2/§3, Catalogue of Attributes** (the `Attribute`/`DataService`/
  `SchemaDistribution` data model and its discovery interfaces) — per TS11
  §1, this part of the spec is designed for hosting by the European
  Commission's Single Digital Gateway and OOTS Semantic Repository
  services, and its discovery interface is specified fully by ETSI TS 119
  478 (OOTS). This is EC-operated infrastructure, distinct from the
  Catalogue of *Attestations* (§4/§5) this registry implements; it is not
  something an issuer-side lab registry stands up. Not previously called
  out explicitly in this document — added here so it isn't rediscovered
  as a gap later.
- **`PUT`/`DELETE /schemas/{schemaId}`** (TS11 §5.3.2/§5.3.3, the
  management API) — already self-declared in
  `internal/httpapi`'s package doc comment; this registry has no
  registration/authorisation workflow (§4.5.3) to gate writes with, since
  entries are added by editing `data/attestations/*.json` and
  redeploying, not via API calls from an external registering entity.
- **JWS-signed `GET` responses** (TS11 §5.3.1) — responses are plain JSON.
  Reasonable to defer for an internal tool with two known consumers over
  a private network path, but worth stating plainly rather than leaving
  implicit.
- **Multi-tenant authorship isolation** (Altia EUD-30 SRS NFR-S-01, FR-05)
  — Altia's sibling implementation scopes type authorship per tenant (a
  tenant, or "EUDIStack" for shared base types), with write isolation as a
  security requirement. This registry has exactly one owner (Fikua) and no
  tenant concept anywhere in `internal/model`; adding one now would model
  for a hypothetical second owner that doesn't exist. Revisit only if this
  registry is ever asked to host definitions on behalf of more than one
  party.
- **CDN / hash-addressed immutable content delivery** (Altia EUD-30 SRS
  NFR-A-01) — Altia serves published artifacts from a CDN under an
  immutable, content-hashed caching policy, decoupling read availability
  from the registry's own uptime. This registry is served directly by
  Traefik at `attestation-registry.fikua.com` with no CDN layer (see
  `CLAUDE.md` "Deployment"). The version-addressed `/type-metadata` route
  (see [Modeling fix: version-addressed type-metadata](#modeling-fix-version-addressed-type-metadata))
  makes responses cacheable in principle, but nothing currently sits in
  front of this service to cache them. Infrastructure-layer change, not a
  domain-model one — revisit if read availability during a deploy or
  outage becomes an actual problem for issuer/verifier.
- **3 of 5 `supportedFormats`/`formatIdentifier` values** (`jwt_vc_json`,
  `jwt_vc_json-ld`, `ldp_vc`) — `model.CredentialFormat` only defines
  `dc+sd-jwt` and `mso_mdoc`. No bundled definition, nor either consumer
  (`fikua-lab-issuer`/`fikua-lab-verifier`), issues or verifies a W3C VC
  format today. Extend `CredentialFormat` the same way `dc+sd-jwt`/
  `mso_mdoc` are defined (`internal/model/enums.go`) if that changes —
  no structural blocker, just no current need.
