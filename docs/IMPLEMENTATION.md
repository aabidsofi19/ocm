# OCM MVP/POC Implementation Notes

This repo implements an MVP/POC for the Operational Complexity Meter (OCM) described in `specs/`.

The current implementation focuses on the MVP metrics called out in `specs/00-overview.md`:

- `CSA` (Configuration Surface Area)
- `DD` (Dependency Depth)

## What You Get

- A single Go binary (`ocm`) that:
  - analyzes a target folder (`--repo`)
  - writes results into SQLite (`--db`)
  - starts an embedded HTTP server with:
    - JSON API (`/api/...`)
    - dashboard UI (`/`)

## How To Run

Build and run:

```bash
go run ./cmd/ocm --repo /path/to/your/repo --db ocm.sqlite --port 8080
```

Then open:

- `http://127.0.0.1:8080/` (dashboard)
- `http://127.0.0.1:8080/api/healthz`

To print the computed result JSON on startup:

```bash
go run ./cmd/ocm --repo . --print
```

## Pipeline Mapping To Specs

The architecture in `specs/00-overview.md` is implemented as:

1) Data sources (scoped MVP)
   - MVP currently scans Kubernetes YAML files (`*.yml`, `*.yaml`) under `--repo`.
   - Helm/Dockerfile/Git metadata are not wired yet in this POC.

2) Parser & normalization (`specs/20-parser-normalization.md`)
   - `internal/parser/yamlscan.go` decodes YAML documents and extracts:
     - env vars
     - resource keys
     - ports
     - replicas
     - a heuristic count of `spec` keys
   - Dependency evidence (heuristic): env var keys ending with `_SERVICE` or `_SERVICE_HOST` with a value that looks like a service host.

3) Metric engine (`specs/30-metric-engine.md`)
   - `CSA` is computed from extracted facts (heuristic approximation).
   - `DD` is computed from a dependency graph:
     - edge direction: `A -> B` means "A depends on B"
     - cycles are handled by condensing strongly-connected components into a DAG and computing longest-path length in the condensed DAG (edge count), per the spec requirement for deterministic behavior.

4) Composite scoring (`specs/40-composite-scoring.md`)
   - Cohort for min/max: all services in the current analysis run.
   - If `max == min`, normalized value is `0`.
   - MVP composite score uses only CSA and DD with default weights `0.5` and `0.5`.
   - Missing metrics are treated as `0` (deterministic).

5) Persistence (`specs/50-persistence-sqlite.md`)
   - `internal/storage/storage.go` manages SQLite schema and writes:
     - `services`
     - `metrics` (CSA/DD)
     - `composite_scores`
   - Timestamp is stored as RFC3339Nano text for stable parsing.

6) Backend API (`specs/60-api-backend.md`)
   - `internal/api/api.go` exposes:
     - `GET /api/healthz`
     - `GET /api/services`
     - `GET /api/services/{id}/scores`
     - `GET /api/services/{id}/metrics/{CSA|DD}`

7) Dashboard (`specs/70-visualization-dashboard.md`)
   - `internal/dashboard/static/` is a no-build dashboard served by the Go binary.
   - Shows:
     - a per-service OCM trend chart
     - latest CSA/DD/OCM values
     - raw JSON panel for debugging

## Service Identification

Spec allows implementation-defined but stable mapping.

This MVP supports a configurable strategy via `--service-key`:

- `dir` (default): service name is the first directory segment under `--repo` that contains the YAML file.
- `manifest`: service name is derived from `metadata.name` (best-effort).

## Known Limitations (POC)

- CSA extraction is heuristic and intentionally simple.
- Dependency extraction is heuristic and based on env-var naming.
- Helm rendering, Dockerfile parsing, and Git change volatility are not implemented yet.
- The analyzer currently runs once on startup; adding a "re-run analysis" API is a natural next step.
