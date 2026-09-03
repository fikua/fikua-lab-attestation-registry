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

Currently defines: EUDI PID (SD-JWT VC), EUDI PID (mdoc), and a non-qualified EAA
prototype for the Barcelona municipal padró (SD-JWT VC).

## Run

```sh
make run          # http://localhost:8080
```

## API

| Method | Path                     | Description                          |
|--------|--------------------------|---------------------------------------|
| GET    | `/api/v1/schemes`        | List all attestation definitions      |
| GET    | `/api/v1/schemes/{id}`   | One definition, by scheme id          |
| GET    | `/healthz`               | Health check                          |
| GET    | `/openapi.yaml`          | OpenAPI spec                          |
| GET    | `/swagger`               | Swagger UI                            |

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
   runs `docker compose pull && up -d`, then polls `/healthz`.

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
