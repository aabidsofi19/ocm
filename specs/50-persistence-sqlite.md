# Persistence Layer (SQLite) - Spec

## Purpose
Persist services, raw metrics, and composite scores as time series for querying and visualization.

## Storage
SQLite (embedded database).

## Schema
From the architecture doc:

### `services`
```
services(
  id INTEGER PRIMARY KEY,
  name TEXT,
  repository TEXT
)
```

### `metrics`
```
metrics(
  id INTEGER PRIMARY KEY,
  service_id INTEGER,
  metric_type TEXT,
  metric_value REAL,
  timestamp DATETIME
)
```

### `composite_scores`
```
composite_scores(
  id INTEGER PRIMARY KEY,
  service_id INTEGER,
  ocm_score REAL,
  timestamp DATETIME
)
```

## Inputs
- Service definitions.
- Metric values per service/timepoint.
- Composite scores per service/timepoint.

## Outputs
- Query results for backend API and/or dashboard:
  - per-service metric history
  - per-service composite score history
  - per-timepoint cross-service comparisons

## Data Integrity
Spec requirements:
- `service_id` MUST refer to an existing row in `services`.
- `metric_type` MUST be a stable identifier (e.g., `CSA`, `DD`, `DB`, `CV`, `FE`, `CDR`).
- Timestamps MUST be consistent across metric and composite score writes for a given run.

## MVP Behavior
- Store CSA and DD metrics and composite scores for 3-5 services.

## Acceptance Criteria
- After a scoring run, the DB contains:
  - one `services` row per service
  - metric rows for computed metrics
  - composite score rows per service/timepoint
