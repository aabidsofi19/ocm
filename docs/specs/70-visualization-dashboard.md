# Visualization Dashboard (Web UI) - Spec

## Purpose
Display operational complexity metrics and trends per service to support interpretable insights and decision-making.

## Technology
Static dashboard assets served by the embedded HTTP server in the single binary.

Implementation notes:
- The UI may be implemented with any frontend stack, but it MUST be built into static files (HTML/CSS/JS) that are embedded into (or shipped alongside) the binary.

## Features (Required)
From the architecture doc:
- Per-service breakdown
- Trend visualization

From the introduction doc (NFR):
- Usability: clear and interpretable visualizations

## Inputs
- Backend-provided data for:
  - services list
  - per-service metric history
  - per-service composite score history

## Outputs
- Interactive views of:
  - OCM score trend per service over time
  - Metric breakdown per service (at least for MVP metrics)

## MVP Requirements
- Display per-service OCM trend over time (explicitly listed in the architecture doc MVP section).
- Display CSA and DD (MVP metrics) per service.

## UI Behavior
Spec requirements:
- Support selecting a service and time window.
- Provide readable labels for metrics (CSA, DD, etc.).

## Error Handling
- Show a clear empty state when no data is available.
- Show a clear error state when the backend is unreachable or returns an error.

## Acceptance Criteria
- With stored results for 3-5 services, the dashboard shows:
  - a per-service trend chart for OCM
  - a per-service breakdown including CSA and DD
