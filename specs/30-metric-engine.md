# Metric Computation Engine - Spec

## Purpose
Compute defined operational complexity metrics per service from normalized facts.

## Inputs
- `ServiceConfigFacts`
- `ServiceDependencyGraph`
- `ServiceChangeFacts`

## Outputs
- Raw metric values per service and timepoint.

## Metrics
The architecture doc defines the following metrics.

### Configuration Surface Area (CSA)
Definition:

`CSA = (# configurable parameters + # environment variables + # resource specifications + # replica configurations)`

Input dependencies:
- Config parsing (Kubernetes/Helm) for parameters/env/resources/replicas.

Output:
- `CSA(service, timestamp) -> number`

### Dependency Depth (DD)
Definition:

`DD = longest path length in the service dependency graph`

Input dependencies:
- Dependency graph extraction.

Spec requirements:
- Define path length as number of edges.
- For graphs with cycles, DD computation MUST be well-defined (e.g., treat cycles as preventing a finite longest path and compute longest simple path in the SCC-condensed DAG). The exact algorithm is implementation-defined but must be deterministic.

Output:
- `DD(service, timestamp) -> number`

### Dependency Breadth (DB)
Definition:

`DB = # direct upstream + downstream dependencies`

Output:
- `DB(service, timestamp) -> number`

### Change Volatility (CV)
Definition:

`CV = # commits affecting service configuration per time window`

Output:
- `CV(service, window_start) -> number`

### Failure Exposure (FE)
Definition:

`FE = # exposed endpoints + external integrations`

Notes:
- The architecture doc defines FE conceptually but does not specify how endpoints/integrations are detected.

Spec requirement:
- FE extraction rules MUST be explicitly documented in the implementation when added.

### Configuration Drift Risk (CDR)
Definition:

`CDR = # environment-specific overrides (dev/staging/prod differences)`

Notes:
- The architecture doc defines CDR conceptually but does not specify override detection rules.

Spec requirement:
- CDR detection rules MUST be explicitly documented in the implementation when added.

## MVP Behavior
For MVP, compute only:
- CSA
- DD

DB/CV/FE/CDR remain defined for future extension.

## Error Handling
- Missing inputs for a metric: produce no metric value for that service/timepoint and record a reason.

## Acceptance Criteria
- Given normalized facts for 3-5 services, returns CSA and DD values per service.
