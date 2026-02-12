# Backend API (Embedded HTTP) - Spec

## Purpose
Provide an embedded HTTP service that orchestrates parsing, metric computation, scoring, persistence, and exposes data for visualization.

## Technology
Go embedded HTTP server shipped in the same single binary as the CLI.

## Responsibilities
- Accept dataset/repository inputs for analysis.
- Execute the pipeline:
  - parse and normalize
  - compute metrics
  - normalize and score
  - persist results to SQLite
- Serve data to the dashboard.

## Inputs
- Repository location (local path) and analysis scope (e.g., commit range/time window).
- Adjustable weights for composite scoring.

## Outputs
- Pipeline execution status.
- Query endpoints for services, metrics, and composite scores.

## API Surface
The architecture doc does not define endpoint contracts.

Spec requirements:
- The implementation MUST document API routes and payloads.
- The API MUST support retrieving per-service breakdowns and trend data required by the dashboard.

## Error Handling
- Validation errors for missing/invalid inputs.
- Structured error reporting for parse failures and pipeline failures.

## Security
Prototype scope:
- No auth requirements are specified in the provided docs.

Spec requirement:
- If the backend is exposed beyond local development, authentication/authorization MUST be added and documented.

## Observability
- Log pipeline start/end, input scope, and counts of extracted services/edges/metrics.
- Log parse failures with file path context.

## Acceptance Criteria
- Can run the end-to-end pipeline on an example repo and serve the resulting time series to the dashboard.
