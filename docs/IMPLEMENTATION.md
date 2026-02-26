# OCM Implementation Notes

This repo implements the Operational Complexity Meter (OCM) described in `specs/`.

All 6 metrics from the specification are implemented:

| Metric | Name | Description |
|--------|------|-------------|
| `CSA` | Configuration Surface Area | Count of configurable knobs (env vars, ports, resource limits, replicas, spec keys) |
| `DD` | Dependency Depth | Longest dependency chain length (SCC-condensed DAG) |
| `DB` | Dependency Breadth | In-degree + out-degree in the service dependency graph |
| `CV` | Change Volatility | Number of commits touching a service's config files within a time window |
| `FE` | Failure Exposure | Externally exposed endpoints and integrations |
| `CDR` | Configuration Drift Risk | Environment-specific config overrides across dev/staging/prod directories |

## What You Get

A single Go binary (`ocm`) that:

- Analyzes a target folder (`--repo`)
- Computes all 6 metrics per service
- Normalizes metrics across the run cohort (min-max)
- Computes a composite OCM score (weighted sum, equal 1/6 weights)
- Writes results into SQLite (`--db`)
- Starts an embedded HTTP server with:
  - JSON API (`/api/...`)
  - Dashboard UI (`/`) with Stripe-inspired design, charts, and metric drilldowns

## How To Run

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

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--repo` | `.` | Path to repo/workspace to analyze |
| `--db` | `ocm.sqlite` | SQLite database file path |
| `--host` | `127.0.0.1` | HTTP server host |
| `--port` | `8080` | HTTP server port |
| `--service-key` | `dir` | Service identity strategy: `dir` or `manifest` |
| `--cv-window` | `30` | Change Volatility lookback window in days |
| `--print` | `false` | Print results as JSON to stdout |

## Pipeline Mapping To Specs

The architecture in `specs/00-overview.md` is implemented as:

### 1. Data Sources (`specs/10-data-sources.md`)

- Kubernetes YAML files (`*.yml`, `*.yaml`) under `--repo`
- Git history via `git log` (for CV metric)
- File system structure (for CDR metric)

### 2. Parser & Normalization (`specs/20-parser-normalization.md`)

`internal/parser/yamlscan.go` decodes YAML documents and extracts:

- **CSA facts**: env vars, resource keys, ports, replicas, spec keys
- **Dependency evidence**: env var keys ending with `_SERVICE` or `_SERVICE_HOST`
- **FE facts**: externally exposed endpoints (see detection rules below)

### 3. Metric Engine (`specs/30-metric-engine.md`)

All 6 metrics are computed in `internal/pipeline/pipeline.go`:

**CSA** — Sum of configuration knobs extracted by the parser.

**DD** — Dependency graph → SCC condensation → longest path (edge count) in the DAG. Cycles are handled deterministically via Tarjan's algorithm.

**DB** — For each service: `in-degree + out-degree` in the dependency graph. In-degree counts how many other services depend on it; out-degree counts its own dependencies.

**CV** — `internal/gitlog/gitlog.go` shells out to `git log --name-only --since=<window>` and counts distinct commits per service. Config extensions: `.yaml`, `.yml`, `.json`, `.toml`, `.env`, `.properties`, `.conf`.

**FE** — Detection rules (documented per spec requirement):

| Signal | Detection | Evidence Component |
|--------|-----------|-------------------|
| Service type LoadBalancer | `spec.type == "LoadBalancer"` → count each port | `lb_port` |
| Service type NodePort | `spec.type == "NodePort"` → count each port | `nodeport` |
| Ingress rules | Each `spec.rules[*].http.paths[*]` path entry | `ingress_path` |
| ExternalName services | `spec.type == "ExternalName"` | `external_integration` |
| Container hostPort | `containers[*].ports[*].hostPort > 0` | `host_port` |
| External URLs in env vars | `http(s)://` URLs excluding `*.svc.cluster.local` | `external_url` |

**CDR** — Detection rules (documented per spec requirement):

Scans for environment-specific directory patterns and counts YAML file overrides across environments. Recognized environment names:
`dev`, `development`, `staging`, `stg`, `production`, `prod`, `qa`, `test`, `uat`, `preview`, `canary`, `local`.

These may appear as top-level directories (`overlays/dev/`), Kustomize-style (`base/` vs `overlays/{dev,prod}/`), or suffixed (`myservice-dev/`). Each unique config file basename appearing in N > 1 environments contributes `N - 1` overrides.

### 4. Composite Scoring (`specs/40-composite-scoring.md`)

- Cohort for min/max normalization: all services in the current analysis run
- If `max == min`, normalized value is `0`
- Composite score: weighted sum of all 6 normalized metrics
- Default weights: equal `1/6` each (configurable via `model.DefaultWeights`)
- Missing metrics are treated as `0`

### 5. Persistence (`specs/50-persistence-sqlite.md`)

`internal/storage/storage.go` manages SQLite schema and writes:
- `services` — service identity
- `metrics` — all metric types (CSA, DD, DB, CV, FE, CDR)
- `composite_scores` — per-service composite OCM scores

### 6. Backend API (`specs/60-api-backend.md`)

`internal/api/api.go` exposes:

- `GET /api/healthz` — health check
- `GET /api/services` — list all services
- `GET /api/overview` — repo-level aggregates for all 6 metrics, per-service summaries
- `GET /api/services/{id}/scores` — composite score history
- `GET /api/services/{id}/metrics/{type}` — metric history for any of the 6 types
- `GET /api/services/{id}/metrics/{type}/evidence` — latest run evidence

### 7. Dashboard (`specs/70-visualization-dashboard.md`)

`internal/dashboard/static/` is a no-build vanilla HTML/CSS/JS dashboard served by the Go binary via `//go:embed`. Features:

- Stripe-inspired dark theme with Inter font
- 6 color-coded metric tiles with sparklines
- Canvas-based OCM trend chart, radar chart (normalized metrics), and bar chart (raw values)
- Service comparison table with inline OCM bars
- Evidence modal for all metric types
- Responsive layout with mobile sidebar toggle

## Service Identification

Configurable strategy via `--service-key`:

- `dir` (default): service name is the first directory segment under `--repo` containing the YAML file
- `manifest`: service name derived from `metadata.name` (best-effort)

## Project Structure

```
cmd/ocm/main.go              CLI entrypoint
internal/
  model/model.go             Metric types, weights, data structures
  parser/yamlscan.go          K8s YAML parser (CSA, FE, dependency extraction)
  pipeline/pipeline.go        Orchestration: all 6 metrics, normalization, composite scoring
  gitlog/gitlog.go            Git history extraction (CV metric)
  storage/storage.go          SQLite persistence
  api/api.go                  HTTP API handlers
  dashboard/
    dashboard.go              Go embed wrapper
    static/                   Embedded HTML/CSS/JS dashboard
specs/                        Specification documents
docs/                         Implementation documentation
```

## Known Limitations

- CSA and dependency extraction are heuristic (env-var naming patterns)
- Helm rendering and Dockerfile parsing are not implemented
- CV requires `git` to be installed and the repo to have history
- CDR detection relies on directory naming conventions
- The analyzer runs once on startup; re-analysis requires restarting the binary
