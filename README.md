# fikua-lab-attestation-registry

Catalogue of Attestations for the Fikua Lab EUDI Wallet ecosystem, per
[ARF 3.0](https://github.com/eu-digital-identity-wallet/eudi-doc-architecture-and-reference-framework)
and [ETSI TS 119 472](https://www.etsi.org/deliver/etsi_ts/119400_119499/).

Standalone Go service (no dependency on `fikua-lab`'s Java backend) that:

- serves the machine-readable Catalogue of Attestations as a JSON API, consumed by
  [`fikua-lab-issuer`](https://github.com/fikua/fikua-lab-issuer) and
  [`fikua-lab-verifier`](https://github.com/fikua/fikua-lab-verifier)
- serves a human-readable Attestation Rulebook browser, since a Rulebook is meant for people
- serves a Swagger UI for the JSON API

Currently defines: EUDI PID (one attestation type, SD-JWT VC and mdoc
formats) and a non-qualified EAA prototype for the Barcelona municipal
padró (SD-JWT VC). See [Compliance](#compliance) below for how these map
to the normative sources.

## Run

```sh
make run          # http://localhost:8080
```

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/schemes` | List all attestation definitions |
| GET | `/api/v1/schemes/{id}` | One definition, by scheme id (this registry's own format) |
| GET | `/api/v1/schemes/{id}/type-metadata` | SD-JWT VC Type Metadata Document for that scheme (404 if it has no `dc+sd-jwt` format) |
| GET | `/health` | Health check |
| GET | `/openapi.yaml` | OpenAPI spec |
| GET | `/swagger` | Swagger UI |

`/api/v1/schemes/{id}` returns this registry's own internal representation
(`AttestationRulebook` + `AttestationScheme`), useful for browsing and for
the human-facing UI. It is **not** a standards-defined credential schema
format. `/api/v1/schemes/{id}/type-metadata` is: it's a real [SD-JWT VC
Type Metadata Document](https://www.ietf.org/archive/id/draft-ietf-oauth-sd-jwt-vc-latest.html#name-sd-jwt-vc-type-metadata),
suitable for a Wallet or Verifier to fetch via the spec's "Registry"
retrieval method — this service acts as that registry for any `vct` it
defines. There is no equivalent generator yet for mdoc (ISO 23220-2
DocType) — see `CLAUDE.md`.

## UI

`/` lists all registered rulebooks; `/rulebooks/{id}` shows one in full detail.

## Build

```sh
make build        # bin/registry, static binary, no CGO
```

Docker image is built `FROM scratch` — no runtime dependencies, single static binary.

## Deployment

CI/CD mirrors `fikua/niu`'s pipeline:

1. `build.yml` — vet/build/test on every push and PR.
2. `release.yml` — on push to `main` or a published release, builds a
   multi-arch image and pushes `docker.io/fikua/fikua-lab-attestation-registry`
   to Docker Hub.
3. `deploy.yml` — manually dispatched, or auto-triggered by a published
   release (gated behind the `prd` GitHub Environment's required
   reviewers). SSHes into the VPS through a Cloudflare Access tunnel,
   syncs `compose.yaml` to `/opt/vps/projects/fikua-lab-attestation-registry/`,
   runs `docker compose pull && up -d`, then polls `/health`.

Public at `https://attestation-registry.fikua.com` — a plain
Cloudflare-proxied A record straight to Traefik, same pattern as `niu` and
`exam-room`. No Cloudflare Access, no Tunnel, no Worker in front (see
`fikua-platform-iac/projects/fikua-lab-attestation-registry/README.md` for
why this deliberately differs from `fikua-lab`'s own
`lab.fikua.com/<role>/` Workers pattern).

Required repo secrets (same as `fikua/niu`): `DOCKER_USERNAME`,
`DOCKER_TOKEN`, `VPS_SSH_PRIVATE_KEY`, `CF_ACCESS_CLIENT_ID`,
`CF_ACCESS_CLIENT_SECRET` (the last two are for tunnelling **SSH access to
the VPS** during deploy, via Cloudflare Access — unrelated to how the
service itself is publicly reached).

## Compliance

[docs/compliance/arf-3.0-etsi-119472-compliance.md](docs/compliance/arf-3.0-etsi-119472-compliance.md)
tracks, requirement by requirement, how the bundled attestation
definitions and domain model map to ARF 3.0, the PID Rulebook, the
Attestation Rulebook template, TS11, ETSI TS 119 472-1, and CIR
2024/2977 — including the gaps and deliberate deviations that remain
(mdoc DocType documents, CDDL-level value constraints, the catalogue-id
vs. lookup-id distinction). Read it before assuming any definition here
is production-ready or before adding a new one — the same requirements
apply.
