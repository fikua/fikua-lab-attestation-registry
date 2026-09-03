# fikua-lab-attestation-registry

Catalogue of Attestations for the Fikua Lab EUDI Wallet ecosystem, per
[ARF 3.0](https://github.com/eu-digital-identity-wallet/eudi-doc-architecture-and-reference-framework)
and [ETSI TS 119 472](https://www.etsi.org/deliver/etsi_ts/119400_119499/).

Standalone Go service (no dependency on `fikua-lab`'s Java backend) that:

- serves the machine-readable Catalogue of Attestations as a JSON API, consumed by
  [`fikua-lab-issuer`](https://github.com/fikua/fikua-lab-issuer) and
  [`fikua-lab-verifier`](https://github.com/fikua/fikua-lab-verifier)
- serves a human-readable Attestation Rulebook browser, since a Rulebook is meant for people

Currently defines: EUDI PID (SD-JWT VC), EUDI PID (mdoc), and a non-qualified EAA
prototype for the Barcelona municipal padró (SD-JWT VC).

## Run

```sh
make run          # http://localhost:8090
```

## API

| Method | Path                     | Description                          |
|--------|--------------------------|---------------------------------------|
| GET    | `/api/v1/schemes`        | List all attestation definitions      |
| GET    | `/api/v1/schemes/{id}`   | One definition, by scheme id          |
| GET    | `/healthz`               | Health check                          |

## UI

`/` lists all registered rulebooks; `/rulebooks/{id}` shows one in full detail.

## Build

```sh
make build        # bin/registry, static binary, no CGO
```

Docker image is built `FROM scratch` — no runtime dependencies, single static binary.
