# Data Sources - Spec

## Purpose
Provide the raw artifacts required to compute OCM metrics for each service, within the scoped MVP sources.

## In Scope (MVP)
- Kubernetes manifests (YAML)
- Helm charts (`values.yaml` + templates)
- Dockerfiles
- Git commit metadata

## Inputs
- One or more repositories containing the above artifacts.
- Repository revision selection (implicit via local checkout or explicit via commit range).

## Outputs
- A set of discovered services and a mapping from service -> source artifacts:
  - paths to Kubernetes/Helm/Dockerfile artifacts
  - commit metadata relevant to each service's configuration

## Service Identification
The system needs a consistent way to map artifacts to a "service".

Specified constraints:
- The architecture document does not define an explicit service discovery standard.

Spec requirement (to implement consistently):
- Service identity MUST be derived deterministically from artifact context (e.g., manifest metadata name and/or directory). The exact mapping is implementation-defined but must be stable across runs.

## Git Commit Metadata
### Required Fields
- commit id (hash)
- timestamp (author/commit time)
- files changed

### Derivations
- Identify commits that affect a service's configuration artifacts (used by Change Volatility).

## Errors and Edge Cases
- Missing or malformed YAML/templates: record parsing errors and skip affected files.
- Repos without Git history or without expected artifacts: treat as empty input, return no services.

## Acceptance Criteria
- Given a repo containing 3-5 services (as suggested for MVP), the system enumerates services and associates them with relevant YAML/Helm/Dockerfile paths and Git commits.
