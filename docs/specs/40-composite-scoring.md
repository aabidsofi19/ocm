# Composite Scoring Engine - Spec

## Purpose
Normalize raw metric values and compute a composite OCM score per service and timepoint using weighted aggregation.

## Inputs
- Raw metric values per service and timepoint.
- Weight vector `w1..w6` (adjustable).

## Outputs
- Normalized metric values per service and timepoint.
- Composite OCM score per service and timepoint.

## Normalization
From the architecture doc:

`Normalized(M) = (M - Min(M)) / (Max(M) - Min(M))`

Spec requirements:
- Min/Max bounds MUST be computed over a defined cohort (e.g., all services in the dataset for the time window). The implementation MUST document the cohort used.
- If `Max(M) == Min(M)`, normalization MUST avoid division by zero; the implementation MUST define the resulting normalized value deterministically (e.g., 0).

## Composite Score
From the architecture doc:

`OCM = w1(CSA) + w2(DD) + w3(DB) + w4(CV) + w5(FE) + w6(CDR)`

Spec requirements:
- Weights MUST be configurable.
- The implementation MUST document default weights.
- Composite score MUST be computed only from metrics available for the given service/timepoint; handling of missing metrics MUST be deterministic (e.g., treat missing as 0 or renormalize weights). The implementation MUST document the chosen approach.

## MVP Behavior
- Compute composite scores from the MVP metrics (CSA, DD) only, using weights applicable to those metrics.

## Acceptance Criteria
- Produces normalized CSA/DD and a composite score for each service/timepoint.
