# fikua-lab-attestation-registry — Project Guide

## Purpose

Catalogue of Attestations for the Fikua Lab EUDI Wallet ecosystem, per ARF 3.0
and ETSI TS 119 472. Standalone service, deliberately separated from
`fikua-lab` (the Java issuer/verifier/wallet-provider monorepo) so that:

- `fikua-lab-issuer` and `fikua-lab-verifier` both consume it with **zero**
  shared code or build dependency on each other or on this registry's
  internals — only the JSON API contract.
- It ships as a single static Go binary (~10-15MB), not a JVM process.
- The Attestation Rulebook is a human-readable document by definition (ARF
  3.0 / TS11 §4.2) — this service therefore serves a UI, not just an API.

## Source of truth

- ARF 3.0: <https://github.com/eu-digital-identity-wallet/eudi-doc-architecture-and-reference-framework>
- Attestation Rulebooks Catalog (upstream template + PID/mDL rulebooks):
  <https://github.com/eu-digital-identity-wallet/eudi-doc-attestation-rulebooks-catalog>
- TS11 (Catalogue of Attributes / Catalogue of Attestations):
  <https://github.com/eu-digital-identity-wallet/eudi-doc-standards-and-technical-specifications/blob/main/docs/technical-specifications/ts11-interfaces-and-formats-for-catalogue-of-attributes-and-catalogue-of-schemes.md>
- ETSI TS 119 472 (issuer metadata / RP presentation / issuance profiles for PID and EAA)

The domain model in `internal/model` mirrors TS11 §4.3 (`SchemaMeta` →
`AttestationScheme`, `Schema` → `FormatSchema`, `TrustAuthority`) and the
Attestation Rulebook template's required fields (`AttestationRulebook`).

## Architecture

```text
cmd/registry/            entrypoint: wiring only, no logic
internal/model/          domain types (Rulebook, Scheme, ClaimDefinition, enums) — no I/O
internal/catalogue/      in-memory registry + JSON loader (embed.FS) + claims validator
internal/httpapi/        JSON API for machine consumers (issuer, verifier)
internal/webui/          human-facing rulebook browser (html/template)
data/attestations/*.json one file per attestation type — the actual catalogue content
web/templates/, web/static/  UI assets, embedded into the binary
```

- **No ORM, no database.** The catalogue is read-only, defined by the JSON
  files under `data/attestations/`, embedded into the binary at build time.
  Adding a credential type = adding a JSON file + rebuilding.
- **No framework.** Standard library only (`net/http`, `html/template`,
  `embed`). Keep it that way unless a real need appears — the whole point of
  this service is to stay small.
- **`internal/model` has zero I/O.** JSON tags only; loading lives in
  `internal/catalogue`.

## Language

- Code, comments, commit messages: English.
- Communication with the user: Catalan, Spanish, or English as they prefer.

## Conventions

- `gofmt` and `go vet` clean before every commit.
- Every new attestation definition needs a test in
  `internal/catalogue/catalogue_test.go` asserting its mandatory claims and
  format-specific encoding, mirroring the pattern already there for PID and
  the padró.
- JSON field values for enums (`presence`, `bindingType`, `attestationLoS`,
  `frameworkType`, `format`, `category`) must match the wire-format strings
  used by TS11 / ETSI TS 119 472, not Go constant names — see
  `internal/model/enums.go` for the exact string values.

## Relationship to fikua-lab

This repo replaces the hardcoded credential-configuration `Map` literals in
`fikua-lab`'s `IssuanceService.buildCredentialConfigurations()`
(`suite/backend/fikua-issuer`). `fikua-lab-issuer` and `fikua-lab-verifier`
are expected to fetch scheme definitions from this service's `/api/v1/schemes`
endpoint at startup or on demand, rather than declaring credential shapes
themselves. Do not duplicate credential/claims definitions back into
`fikua-lab` — this registry is the single source of truth.
