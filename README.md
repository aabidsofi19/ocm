# OCM (Operational Complexity Meter) – MVP/POC

This is an MVP/POC implementation of the Operational Complexity Meter described in `specs/`.

Current MVP scope:

- Parses Kubernetes YAML (`*.yml` / `*.yaml`) under a target folder
- Computes MVP metrics:
  - `CSA` (Configuration Surface Area)
  - `DD` (Dependency Depth)
- Normalizes metrics across the run cohort and computes a composite score (CSA/DD only)
- Stores timepoints into SQLite
- Serves an embedded dashboard + JSON API

## Run

```bash
go run ./cmd/ocm --repo /path/to/repo --db ocm.sqlite --port 8080
```

Open `http://127.0.0.1:8080/`.

## Docs

- Specs: `specs/`
- Implementation notes: `docs/IMPLEMENTATION.md`
