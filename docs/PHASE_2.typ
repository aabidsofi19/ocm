// Study Project – Phase 2 Document (Typst)


#set page(margin: (top: 22mm, bottom: 22mm, left: 20mm, right: 20mm))
#set text(size: 11pt)
#set heading(numbering: "1.")


// =====================================================
// Cover Page
// =====================================================

#align(center)[
  #text(size: 20pt, weight: 700)[Study Project – Phase 2 Document]
  #text(size: 13pt)[(Design & Proof of Concept)]
]

#v(16mm)

#text(weight: 700)[Cover Page]

#v(4mm)

#grid(
  columns: (38%, 62%),
  row-gutter: 6pt,
  column-gutter: 10pt,
  [Course Title:], [Operational Complexity Meter (OCM): Quantifying Operational Complexity in Distributed Systems],
  [Project Title:], [Operational Complexity Meter (OCM) – MVP/POC],
  [Group Number:], [1],
  [Student Name(s):], [Aabid Ali Sofi],
  [Student ID(s):], [2023EBCS041],
  [Project Advisor / Supervisor:], [Preethy P Johny],
  [Date of Submission:], [Feb 12, 2026],
)

#pagebreak()


// =====================================================
// 1. Introduction
// =====================================================

= Introduction

== Purpose of Phase 2

Phase 2 translates the Phase 1 idea (quantifying operational complexity in distributed systems) into an implementable system design and a working proof of concept.

This phase focuses on:

- Converting the problem definition into a concrete architecture and requirements.
- Defining the MVP scope and module boundaries for the prototype system.
- Validating feasibility by implementing an end-to-end PoC that computes and explains operational complexity metrics.

== Scope of Phase 2

Included:

- System architecture.
- Functional and non-functional requirements for the MVP.
- Database schema and data flow design.
- A working PoC implementation (single-binary Go app) that:
  - Parses Kubernetes YAML
  - Computes metrics
  - Stores results in SQLite
  - Serves a local web UI and JSON API
  - Supports evidence drill-down for the CSA metric

Excluded (prototype scope):

- Production-scale ingestion and scheduled execution.
- Authentication/authorization and multi-user access.
- Advanced data sources (Helm rendering, Dockerfile parsing, Git-based volatility) beyond what is implemented in the PoC.


// =====================================================
// 2. System Overview
// =====================================================

= System Overview

== Product Perspective

OCM is a standalone prototype tool used by engineers to analyze a repository/workspace containing deployment artifacts (Kubernetes YAML).

In the PoC implementation, OCM is shipped as a single executable that:

- Runs locally against a repo path (`--repo`)
- Persists results into a local SQLite DB (`--db`)
- Serves a local dashboard and API over HTTP

Deployment environment:

- Local development machine
- Local HTTP server (localhost usage)

== Major System Functions

- Discover Kubernetes YAML manifests under a repository path
- Identify services deterministically
- Parse and normalize configuration facts
- Compute MVP metrics:
  - CSA (Configuration Surface Area)
  - DD (Dependency Depth)
- Normalize metrics and compute a composite OCM score
- Persist services, metrics, scores, and evidence into SQLite
- Provide a dashboard for aggregates, per-service values, and drill-down

== User Classes

- Engineers / SRE / DevOps (primary users)
- Architects / Tech Leads
- No external integrations in PoC


// =====================================================
// 3. Functional Requirements
// =====================================================

= Functional Requirements

FR1: The system shall accept a local repository/workspace path as input.

FR2: The system shall discover Kubernetes manifests (`.yml` / `.yaml`) under the input path.

FR3: The system shall deterministically map discovered artifacts to a stable service identity.

FR4: The system shall parse Kubernetes YAML and extract configuration facts required to compute CSA.

FR5: The system shall compute CSA per service for a given analysis run.

FR6: The system shall compute DD per service from an extracted dependency graph, with deterministic behavior in the presence of cycles.

FR7: The system shall normalize metric values across a defined cohort and compute a composite OCM score.

FR8: The system shall persist services, metric values, composite scores, and metric evidence into a SQLite database.

FR9: The system shall expose a local HTTP API for reading:

- services
- metric time series
- composite score time series
- repo-level aggregates
- metric evidence for the latest run

FR10: The system shall serve a local dashboard that displays:

- repo CSA aggregate and repo OCM aggregate
- per-service latest CSA and OCM
- evidence drill-down for CSA


// =====================================================
// 4. Non-Functional Requirements
// =====================================================

= Non-Functional Requirements

== Performance

- Analysis completes within seconds to minutes for small-to-medium repos.
- API and dashboard responses are sub-second.

== Security

- Localhost-only usage in PoC.
- Authentication required for production usage.

== Usability

- Clear metric labels and error states.
- Evidence drill-down shows file path and manifest context.

== Maintainability

- Modular package structure.
- Extensible metrics architecture.


// =====================================================
// 5. System Architecture and Design
// =====================================================

= System Architecture and Design

== System Architecture Diagram

#figure(
  image("./system-arch-phase2.png"),
  caption: [High-level OCM pipeline (MVP/POC)],
)<system-architecture>

Brief explanation:

- CLI triggers analysis.
- Metrics computed and evidence attached.
- Results persisted into SQLite.
- Embedded server exposes API and dashboard.


== Implementation references are based on the current PoC code:

- CLI + Server (`cmd/ocm/main.go`)
  - Runs the analysis pipeline once on startup
  - Migrates the DB schema
  - Saves the run results
  - Starts HTTP server for `/api/*` + `/` dashboard

- Parser & Normalization (`internal/parser/yamlscan.go`)
  - Inputs: Kubernetes YAML documents
  - Outputs: per-document facts for CSA, inferred dependencies, evidence items
  - Evidence includes: component type (env/port/resource/replica/spec_key), key/value, source file path, manifest kind/name

- Pipeline (`internal/pipeline/pipeline.go`)
  - Discovers YAML files
  - Service identification strategies:
    - `dir` (default): service name = first directory segment under `--repo`
    - `manifest`: service name from `metadata.name`
  - Aggregates parsed facts per service
  - Computes metrics (CSA, DD)
  - Normalizes and computes composite OCM score

- Metric Engine
  - CSA: computed as a count of extracted configuration facts
  - DD: computed as longest path length in the SCC-condensed DAG (cycles handled deterministically)

- Storage (SQLite) (`internal/storage/storage.go`)
  - Migrates schema
  - Saves per-run metrics, scores, and evidence
  - Serves query methods for API

- Embedded API (`internal/api/api.go`)
  - Provides health, service listing, metric and score series, overview aggregates, and evidence endpoints

- Dashboard (`internal/dashboard/`)
  - Static HTML/CSS/JS assets embedded into the Go binary
  - Provides per-service selection and metric drill-down modal

== Data Flow Design

End-to-end data flow in the PoC:

1) Input selection
   - User runs the CLI with `--repo` and `--db`.

2) Discovery
   - Walk the repo filesystem and collect `.yml`/`.yaml` files.

3) Parsing and fact extraction
   - Decode YAML documents
   - Extract CSA contributors and store them as evidence
   - Extract inferred dependencies (best-effort heuristics)

4) Metric computation
   - CSA is aggregated per service
   - DD is computed from the inferred dependency graph

5) Normalization and composite
   - Cohort: all services in the current run
   - Normalization: `(M - min) / (max - min)`; if `max == min`, normalized value is `0`
   - Composite score: MVP uses CSA and DD only (default weights 0.5/0.5)

6) Persistence
   - Insert/ensure service rows
   - Insert metric rows
   - Insert composite score row
   - Insert metric evidence rows

7) Presentation
   - UI loads `/api/overview` and `/api/services`
   - UI loads per-service metric series and score series
   - UI loads metric evidence on click


// =====================================================
// Database Design
// =====================================================

== Database Design

Core tables:

```sql
services(
  id INTEGER PRIMARY KEY,
  name TEXT,
  repository TEXT
);

metrics(
  id INTEGER PRIMARY KEY,
  service_id INTEGER,
  metric_type TEXT,
  metric_value REAL,
  timestamp TEXT
);

composite_scores(
  id INTEGER PRIMARY KEY,
  service_id INTEGER,
  ocm_score REAL,
  timestamp TEXT
);
```

Evidence table:

```sql
metric_evidence(
  id INTEGER PRIMARY KEY,
  service_id INTEGER,
  metric_type TEXT,
  component TEXT,
  evidence_key TEXT,
  evidence_value TEXT,
  source_path TEXT,
  manifest_kind TEXT,
  manifest_name TEXT,
  timestamp TEXT
);
```


// =====================================================
// 6. Technology Stack
// =====================================================

= Technology Stack

- Backend/CLI: Go
  - Single-binary distribution and a simple local deployment story.
  - Good fit for embedded HTTP server and fast file scanning.

- Database: SQLite (embedded)
  - Zero configuration and deterministic storage.
  - Easy to query for dashboard needs.

- Parsing: `gopkg.in/yaml.v3`
  - Reliable YAML decoding for Kubernetes manifests.

- UI: Static HTML/CSS/JS
  - Served directly by the embedded server.
  - No runtime frontend build chain required.




// =====================================================
// 7. Proof of Concept
// =====================================================

= Proof of Concept (PoC)

== PoC Demo Video 
#link("https://youtu.be/IMpt782NJFI")[
  ▶ Watch Demo Video
]


== PoC Run Command



```bash
just run REPO=/path/to/repo DB=ocm.sqlite PORT=8080 SERVICE_KEY=dir
```

Dashboard URL:

- http://127.0.0.1:8080/

== Demonstrated Features

- Repo-level CSA and OCM aggregates
- Per-service latest values
- CSA evidence drill-down
- REST API endpoints for metrics and scores

Current limitations:

- CSA extraction heuristic
- Dependency extraction heuristic
- No Helm or Git integration
- Runs only on startup


// =====================================================
// 8. Testing
// =====================================================

= Testing and Validation

- Unit tests for DD cycles and normalization edge cases
- Manual validation using example repositories
- Evidence verification against YAML sources


// =====================================================
// 9. Risks
// =====================================================

= Risks and Mitigation

R1: Service identification ambiguity  
Mitigation: Multiple strategies (`dir`, `manifest`)

R2: Dependency extraction uncertainty  
Mitigation: Deterministic cycle handling

R3: Metric validity vs heuristics  
Mitigation: Transparent evidence storage

R4: Scope creep  
Mitigation: Strict MVP focus (CSA, DD)


// =====================================================
// 10. Outcomes
// =====================================================

= Phase 2 Outcomes

- Concrete architecture aligned with specs
- Working end-to-end PoC
- Metric computation, persistence, visualization
- Evidence-backed explainability

== Readiness for Phase 3

- Modular codebase ready for:
  - Additional metrics (CV/FE/CDR)
  - Helm and Git integration
  - Enhanced UI


// =====================================================
// 11. Supervisor Review
// =====================================================

= Supervisor Review and Approval

Advisor Feedback:

Supervisor Comments:

Recommendations:

Signature: 

Date: 

