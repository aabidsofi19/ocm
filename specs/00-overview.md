# Operational Complexity Meter (OCM) - System Overview Spec

## Purpose
Operational Complexity Meter (OCM) defines and quantifies "operational complexity" as a measurable system property for distributed systems. It aggregates multiple operational dimensions (configuration, dependencies, change volatility, failure exposure, drift risk) into a composite score so teams can track operational risk over time and across services.

## Scope
This spec covers a prototype architecture (Phase 2) and the MVP scope described in the architecture document.

Included:
- Ingest configuration, dependency, and change data from scoped sources.
- Compute operational complexity sub-metrics.
- Normalize metrics and compute a composite OCM score via weighted aggregation.
- Store per-service metrics and time-series scores.
- Visualize per-service breakdowns and trends.

Excluded:
- Automated remediation/optimization.
- Large-scale production validation.

## Key Concepts
### Service
A logical unit being scored (typically a microservice), identified by name and optionally repository.

### Metric
A numeric value representing one operational complexity dimension.

### Composite Score
A weighted sum of normalized metrics, intended as a single operational complexity indicator.

## Architecture (Proposed)
Data sources -> Parser & Normalization Layer -> Metric Engine -> Composite Scoring Engine -> SQLite Database -> Visualization Dashboard.

### Data Sources (Scoped for MVP)
- Kubernetes manifests (YAML)
- Helm charts (`values.yaml` + templates)
- Dockerfiles
- Git commit metadata

## Operational Metrics (Defined)
1. Configuration Surface Area (CSA)
   CSA = (Number of configurable parameters + Environment variables + Resource specifications + Replica configurations)

2. Dependency Depth (DD)
   DD = longest path length in the service dependency graph

3. Dependency Breadth (DB)
   DB = number of direct upstream + downstream dependencies

4. Change Volatility (CV)
   CV = number of commits affecting service configuration per time window

5. Failure Exposure (FE)
   FE = number of exposed endpoints + external integrations

6. Configuration Drift Risk (CDR)
   CDR = count of environment-specific overrides (dev/staging/prod differences)

### Normalization
All metrics are normalized between 0 and 1:

`Normalized(M) = (M - Min(M)) / (Max(M) - Min(M))`

### Composite OCM
`OCM = w1(CSA) + w2(DD) + w3(DB) + w4(CV) + w5(FE) + w6(CDR)`

Weights `w1..w6` are adjustable.

## MVP Scope
For the MVP implementation, only the following metrics are computed:
- CSA
- DD

Rationale (from the architecture doc): measurable using static configuration + Git data; demonstrates structural and temporal complexity; feasible within semester constraints.

## Non-Functional Requirements
From the introduction doc:
- Usability: clear, interpretable visualizations
- Performance: efficient computation on moderate datasets
- Scalability: support small-to-medium distributed systems
- Maintainability: modular metric definitions

## Technology Stack (Updated)
Implementation target: a single self-contained binary.

- Language: Go
- Packaging: single executable containing CLI + embedded HTTP server
- DB: SQLite (embedded)
- Parsing:
  - Kubernetes YAML: Go YAML + (optional) Kubernetes API object decoding
  - Helm charts: Helm v3 Go SDK rendering (no external `helm` required)
  - Dockerfiles: Go Dockerfile parser (or minimal parsing for required fields)
  - Git metadata: Go Git library (or optional shell-out to `git` when available)
- Data processing: in-process computation (no Pandas dependency)
- UI: static dashboard assets served by the embedded HTTP server (framework is build-time only if used)

## Assumptions and Constraints
- Simplified dependency models.
- Limited access to real-world operational data.
- No dependency on production environments.
