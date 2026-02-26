# OCM (Operational Complexity Meter)

Scans Kubernetes YAML repositories, computes 6 operational complexity metrics per service, stores results in SQLite, and serves a web dashboard.

## Metrics

| Metric | Description |
|--------|-------------|
| **CSA** | Configuration Surface Area -- configurable knobs count |
| **DD** | Dependency Depth -- longest dependency chain |
| **DB** | Dependency Breadth -- total direct dependencies (in + out) |
| **CV** | Change Volatility -- config commits in a time window |
| **FE** | Failure Exposure -- externally exposed endpoints |
| **CDR** | Configuration Drift Risk -- env-specific config overrides |

A composite OCM score (0-1) is computed as a weighted sum of all normalized metrics.

## Run

```bash
go run ./cmd/ocm --repo /path/to/repo --db ocm.sqlite --port 8080
```

Open `http://127.0.0.1:8080/`.

### Flags

```
--repo          Path to repo to analyze (default: .)
--db            SQLite DB file (default: ocm.sqlite)
--host          HTTP host (default: 127.0.0.1)
--port          HTTP port (default: 8080)
--service-key   Service identity: dir|manifest (default: dir)
--cv-window     CV lookback window in days (default: 30)
--print         Print results as JSON to stdout
```

## Docs

- Specs: `specs/`
- Implementation notes: `docs/IMPLEMENTATION.md`
