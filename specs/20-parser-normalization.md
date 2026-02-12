# Parser & Normalization Layer - Spec

## Purpose
Parse the scoped source artifacts and normalize extracted information into a consistent internal representation for downstream metric computation.

## Submodules
From the architecture doc:
- YAML/JSON parser
- Dependency graph extractor
- Change frequency analyzer

This document specifies the parsing and normalization responsibilities; metric formulas are specified in `docs/specs/30-metric-engine.md`.

## Inputs
- Kubernetes YAML manifests
- Helm charts (`values.yaml` + templates)
- Dockerfiles
- Git commit metadata

## Implementation Target
Single-binary Go implementation.

Spec requirements:
- Helm templates SHOULD be rendered using the Helm v3 Go SDK so the tool does not require an external `helm` binary.
- Git history MAY be read using a Go library or by shelling out to `git` when available; behavior MUST be documented and deterministic.

## Outputs
Normalized representations:
- `ServiceConfigFacts`: per-service facts extracted from config (parameters, env vars, replicas, ports, resource limits)
- `ServiceDependencyGraph`: directed graph of service dependencies
- `ServiceChangeFacts`: per-service time series counts of config-changing commits per time window

## Configuration Parsing (Kubernetes + Helm)
### Extracted Fields (explicitly listed in architecture doc)
- configurable parameters
- environment variables
- replicas
- ports
- resource limits/specifications

### Normalization Rules
- Represent extracted config as key/value facts with stable keys.
- Preserve the source path and (if applicable) the manifest kind/name for traceability.

## Dependency Graph Extraction
### Purpose
Construct a service dependency graph from manifests.

### Evidence Sources (from architecture doc examples)
- service-to-service references
- ingress rules

### Graph Model
- Node: service
- Edge: A -> B indicates A depends on B (dependency direction MUST be consistent across all computations)

### Required Output
- Adjacency list or equivalent structure per revision/timepoint.

## Change Frequency Analysis
### Purpose
Compute how frequently service configuration changes over time using Git history.

### Required Output
- For each service: count of commits that affect the service's configuration artifacts per time window.

### Time Window
The architecture doc defines "per time window" but does not specify the window size.

Spec requirement:
- The time window MUST be configurable (e.g., daily/weekly), with a documented default.

## Errors and Edge Cases
- Incomplete dependency evidence: graph is best-effort and may be sparse.
- Cycles: graph may contain cycles; downstream computations MUST define behavior (see DD spec).
- Template rendering failures (Helm): if templates cannot be rendered, parsing should fall back to values and any directly parseable content.

## Observability
- Record parsing error counts by artifact type and path.
- Record number of services discovered and number of edges extracted.

## Acceptance Criteria
- Extracts the fields listed in the architecture doc from example Kubernetes/Helm artifacts.
- Produces a dependency graph usable to compute longest path length.
- Produces per-service commit counts per configured time window.
