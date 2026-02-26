// OCM — Operational Complexity Meter
// Full Project Presentation (Polylux)

#import "@preview/polylux:0.4.0": *

#set page(paper: "presentation-16-9")
#set text(size: 20pt, font: "Inter")

// ─── Color palette (matches dashboard design tokens) ───
#let bg      = rgb("#0a0d12")
#let surface = rgb("#111318")
#let raised  = rgb("#181b22")
#let border  = rgb("#2a2e38")
#let txt     = rgb("#f0f0f3")
#let txt2    = rgb("#9198a1")
#let txt3    = rgb("#6b727d")
#let accent  = rgb("#635bff")
#let csa-c   = rgb("#635bff")
#let dd-c    = rgb("#0ea5e9")
#let db-c    = rgb("#8b5cf6")
#let cv-c    = rgb("#f59e0b")
#let fe-c    = rgb("#ef4444")
#let cdr-c   = rgb("#10b981")

// ─── Reusable helpers ───
#let title-slide(title, subtitle) = slide[
  #set page(fill: bg)
  #set text(fill: txt)
  #set align(horizon + center)
  #block(width: 100%)[
    #text(size: 18pt, fill: accent, weight: 600, tracking: 0.12em)[STUDY PROJECT — PHASE 3]
    #v(12pt)
    #text(size: 44pt, weight: 700)[#title]
    #v(8pt)
    #text(size: 20pt, fill: txt2)[#subtitle]
    #v(24pt)
    #line(length: 120pt, stroke: 1.5pt + accent)
    #v(16pt)
    #text(size: 14pt, fill: txt3)[
      Aabid Ali Sofi #h(12pt) | #h(12pt) 2023EBCS041 \
      Advisor: Preethy P Johny #h(12pt) | #h(12pt) Feb 2026
    ]
  ]
]

#let section-slide(title) = slide[
  #set page(fill: surface)
  #set text(fill: txt)
  #set align(horizon + center)
  #block(width: 100%)[
    #text(size: 14pt, fill: accent, weight: 600, tracking: 0.1em)[SECTION]
    #v(8pt)
    #text(size: 38pt, weight: 700)[#title]
    #v(12pt)
    #line(length: 80pt, stroke: 1.5pt + accent)
  ]
]

#let content-slide(title, body) = slide[
  #set page(fill: bg)
  #set text(fill: txt)
  #v(8pt)
  #text(size: 28pt, weight: 700)[#title]
  #v(4pt)
  #line(length: 100%, stroke: 0.5pt + border)
  #v(8pt)
  #text(size: 17pt, fill: txt2)[#body]
]

#let metric-pill(label, color) = box(
  fill: color.lighten(85%),
  radius: 4pt,
  inset: (x: 8pt, y: 3pt),
  text(size: 13pt, weight: 600, fill: color)[#label]
)

// ═══════════════════════════════════════════════════════
//  SLIDES
// ═══════════════════════════════════════════════════════

// ─── 1. Title ───
#title-slide(
  [Operational Complexity Meter],
  [Quantifying Operational Complexity in Distributed Systems]
)

// ─── 2. Agenda ───
#content-slide([Agenda])[
  #set text(size: 18pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 24pt,
    [
      + Problem Statement
      + What is OCM?
      + The 6 Metrics
      + System Architecture
      + Pipeline Deep Dive
    ],
    [
      6. Parser & Extraction
      7. Graph Algorithms
      8. Storage & API
      9. Dashboard
      10. Demo & Results
      11. Future Work
    ]
  )
]

// ─── 3. Problem Statement ───
#section-slide([Problem Statement])

#content-slide([The Problem])[
  Modern distributed systems are *operationally complex*, but that complexity is
  rarely measured.

  #v(12pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      *Symptoms:*
      - Hundreds of config parameters
      - Deep dependency chains
      - Frequent config changes
      - Env-specific drift (dev vs prod)
      - Exposed failure surfaces
    ],
    [
      *Impact:*
      - Outages from config errors
      - Slow incident response
      - Hidden coupling between services
      - Unmeasured toil & risk
    ]
  )

  #v(12pt)
  #align(center)[
    #box(fill: accent.lighten(90%), radius: 6pt, inset: 12pt)[
      #text(fill: accent, weight: 600)[Can we quantify operational complexity as a single, measurable score?]
    ]
  ]
]

// ─── 4. What is OCM? ───
#section-slide([What is OCM?])

#content-slide([Operational Complexity Meter])[
  #set text(size: 17pt)

  OCM is a *single Go binary* that:

  #v(8pt)

  #grid(
    columns: (auto, 1fr),
    column-gutter: 12pt,
    row-gutter: 10pt,
    text(fill: accent, weight: 700)[1.], [*Scans* a Kubernetes YAML repository],
    text(fill: accent, weight: 700)[2.], [*Identifies* services from directory structure or manifest metadata],
    text(fill: accent, weight: 700)[3.], [*Computes* 6 operational complexity metrics per service],
    text(fill: accent, weight: 700)[4.], [*Normalizes* and produces a composite OCM score (0–1)],
    text(fill: accent, weight: 700)[5.], [*Stores* results in SQLite for time-series tracking],
    text(fill: accent, weight: 700)[6.], [*Serves* an interactive web dashboard],
  )

  #v(12pt)

  ```bash
  go run ./cmd/ocm --repo /path/to/k8s-repo --db ocm.sqlite --port 8080
  ```

  #v(6pt)
  #text(size: 14pt, fill: txt3)[
    No external dependencies. No K8s SDK. No CGO. Pure Go.
  ]
]

// ─── 5. The 6 Metrics ───
#section-slide([The 6 Metrics])

#content-slide([Metric Overview])[
  #set text(size: 16pt)
  #table(
    columns: (auto, 1fr, 1fr),
    align: (left, left, left),
    stroke: 0.5pt + border,
    inset: 8pt,
    fill: (x, y) => if y == 0 { raised } else { none },

    [*Metric*], [*What It Measures*], [*How*],

    [#metric-pill("CSA", csa-c)], [Configuration Surface Area], [Count of env vars, ports, resources, replicas, spec keys],
    [#metric-pill("DD", dd-c)], [Dependency Depth], [Longest path in the SCC-condensed dependency DAG],
    [#metric-pill("DB", db-c)], [Dependency Breadth], [Direct upstream + downstream dependency count],
    [#metric-pill("CV", cv-c)], [Change Volatility], [Git commits affecting config in the last 30 days],
    [#metric-pill("FE", fe-c)], [Failure Exposure], [Exposed endpoints (LB, NodePort, Ingress, hostPort, ext URLs)],
    [#metric-pill("CDR", cdr-c)], [Config Drift Risk], [Env-specific overrides (dev/staging/prod directory diffs)],
  )
]

#content-slide([Composite OCM Score])[
  #set text(size: 18pt)

  *Normalization* — Min-max across the service cohort:

  #v(6pt)
  #align(center)[
    $ "Normalized"(M) = (M - M_"min") / (M_"max" - M_"min") $
  ]

  #v(12pt)

  *Composite Score* — Weighted sum of normalized metrics:

  #v(6pt)
  #align(center)[
    $ "OCM" = 1/6 dot "CSA" + 1/6 dot "DD" + 1/6 dot "DB" + 1/6 dot "CV" + 1/6 dot "FE" + 1/6 dot "CDR" $
  ]

  #v(12pt)

  - Output range: *0.0* (minimal complexity) to *1.0* (maximum in cohort)
  - Weights are *configurable* (default: equal 1/6 each)
  - If max = min for a metric, normalized value = 0 (no variance)
]

// ─── 6. Architecture ───
#section-slide([System Architecture])

#content-slide([Single-Binary Architecture])[
  #set text(size: 16pt)
  #align(center)[
    #block(fill: raised, radius: 8pt, inset: 16pt, width: 90%)[
      #text(size: 14pt, fill: txt2)[
        ```
        ┌─────────────────────────────────────────────────────────┐
        │                    ocm (single binary)                  │
        │                                                         │
        │  ┌──────────┐   ┌──────────┐   ┌────────────────────┐  │
        │  │ CLI/Main │──▶│ Pipeline │──▶│  SQLite Storage    │  │
        │  └──────────┘   │          │   └────────────────────┘  │
        │                 │ Parser   │            │              │
        │                 │ GitLog   │            ▼              │
        │                 │ Metrics  │   ┌────────────────────┐  │
        │                 │ Scoring  │   │  HTTP API Server   │  │
        │                 └──────────┘   │  + Embedded        │  │
        │                                │    Dashboard       │  │
        │                                └────────────────────┘  │
        └─────────────────────────────────────────────────────────┘
        ```
      ]
    ]
  ]

  #v(8pt)
  #grid(
    columns: (1fr, 1fr, 1fr),
    column-gutter: 16pt,
    [
      *Only 2 dependencies:*
      - `yaml.v3`
      - `modernc.org/sqlite`
    ],
    [
      *No CGO required* \
      Pure Go SQLite driver \
      Cross-compilable
    ],
    [
      *Embedded assets* \
      `go:embed` for HTML/CSS/JS \
      Zero external files
    ]
  )
]

#content-slide([Data Flow Pipeline])[
  #set text(size: 17pt)
  #grid(
    columns: (auto, 1fr),
    column-gutter: 12pt,
    row-gutter: 14pt,
    text(fill: accent, weight: 700)[1.], [*Discovery* — Walk filesystem, collect `.yaml`/`.yml`, skip `.git`/`vendor`],
    text(fill: accent, weight: 700)[2.], [*Parsing* — Multi-doc YAML decode, extract CSA facts + deps + FE signals],
    text(fill: accent, weight: 700)[3.], [*Service ID* — Map files to services via `dir` (first path component) or `manifest` (metadata.name)],
    text(fill: accent, weight: 700)[4.], [*Metric Engine* — Compute DD (graph), DB (degree), CV (git), FE (exposure), CDR (drift)],
    text(fill: accent, weight: 700)[5.], [*Normalization* — Min-max across cohort for each metric],
    text(fill: accent, weight: 700)[6.], [*Composite* — Weighted sum → single OCM score per service],
    text(fill: accent, weight: 700)[7.], [*Persist* — Atomic transaction: services + metrics + evidence + scores → SQLite],
    text(fill: accent, weight: 700)[8.], [*Serve* — HTTP API + embedded dashboard at `localhost:8080`],
  )
]

// ─── 7. Parser Deep Dive ───
#section-slide([Parser & Extraction])

#content-slide([CSA Extraction])[
  #set text(size: 17pt)

  For each Kubernetes manifest, count configuration surface contributors:

  #v(8pt)

  #table(
    columns: (auto, 1fr, auto),
    stroke: 0.5pt + border,
    inset: 8pt,
    fill: (x, y) => if y == 0 { raised } else { none },
    [*Component*], [*Source*], [*Count*],
    [`env`], [Container environment variables], [1 per var],
    [`port`], [Container port entries], [1 per port],
    [`resource`], [Resource limits/requests leaf keys], [1 per key],
    [`replica`], [`spec.replicas` field], [1 if present],
    [`spec_key`], [Top-level keys under `spec`], [1 per key],
  )

  #v(8pt)

  Each contributor is stored as an *evidence item* with source path, manifest kind, and manifest name for full traceability.
]

#content-slide([Dependency Extraction])[
  #set text(size: 16pt)

  Three heuristic detection rules for inter-service dependencies:

  #v(8pt)

  *Rule 1 — Env var suffix matching:*
  #text(size: 14pt)[
    `_SERVICE`, `_SERVICE_HOST`, `_SERVICE_PORT`, `_ADDR`, `_HOST`, `_URL`, `_ENDPOINT`, `_URI`
  ]

  #v(6pt)
  *Rule 2 — Cluster FQDN detection:*
  #text(size: 14pt)[
    Values containing `.svc.cluster.local` → extract service name
  ]

  #v(6pt)
  *Rule 3 — Hostname:port pattern:*
  #text(size: 14pt)[
    Regex: `^[a-z][a-z0-9-]{0,62}:\d{2,5}$` → extract hostname
  ]

  #v(12pt)
  #text(size: 14pt, fill: txt3)[
    Designed for real-world repos: Sock Shop, Google Online Boutique, Helm charts, Rails/Django/Spring deployments.
  ]
]

#content-slide([Failure Exposure Detection])[
  #set text(size: 16pt)

  FE counts exposed endpoints and external integrations — 5 detection rules:

  #v(8pt)

  #grid(
    columns: (auto, 1fr),
    column-gutter: 12pt,
    row-gutter: 10pt,
    text(fill: fe-c, weight: 700)[1.], [*LoadBalancer / NodePort* service types — each port = 1 endpoint],
    text(fill: fe-c, weight: 700)[2.], [*Ingress* resources — each `rules[].http.paths[]` = 1 endpoint],
    text(fill: fe-c, weight: 700)[3.], [*ExternalName* services — 1 external integration],
    text(fill: fe-c, weight: 700)[4.], [*Container hostPort* — each hostPort = 1 endpoint],
    text(fill: fe-c, weight: 700)[5.], [*External URLs* in env vars — `http(s)://` not containing `.svc.cluster.local`],
  )

  #v(12pt)
  #text(size: 14pt, fill: txt3)[
    Detection rules are explicitly documented in code per spec requirement.
  ]
]

// ─── 8. Graph Algorithms ───
#section-slide([Graph Algorithms])

#content-slide([Dependency Depth — Tarjan's SCC])[
  #set text(size: 16pt)

  DD = longest path (edge count) in the dependency graph.

  #v(8pt)

  *Problem:* Dependency graphs may have cycles (A → B → A).

  *Solution:* SCC (Strongly Connected Components) condensation:

  #v(6pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      *Step 1:* Tarjan's algorithm finds SCCs \
      *Step 2:* Condense each SCC to a single node \
      *Step 3:* Result is a DAG (no cycles) \
      *Step 4:* DFS with memoization → longest path
    ],
    [
      *Example:*
      ```
      A → B → C    (B ↔ A = cycle)
          ↑ ↙
          A

      SCC: {A, B} → node S₁
      DAG: S₁ → C
      DD[A] = DD[B] = 1
      DD[C] = 0
      ```
    ]
  )

  #v(8pt)
  #text(size: 14pt, fill: txt3)[
    Algorithm is deterministic — same input always produces same DD values.
  ]
]

// ─── 9. Storage & API ───
#section-slide([Storage & API])

#content-slide([SQLite Schema])[
  #set text(size: 15pt)

  4 tables, atomic transactions, time-series support:

  #v(6pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 16pt,
    [
      *Core tables:*
      ```sql
      services(id, name, repository)
        UNIQUE(name, repository)

      metrics(id, service_id,
        metric_type, metric_value,
        timestamp)

      composite_scores(id, service_id,
        ocm_score, timestamp)
      ```
    ],
    [
      *Evidence table:*
      ```sql
      metric_evidence(id, service_id,
        metric_type, component,
        evidence_key, evidence_value,
        source_path,
        manifest_kind, manifest_name,
        timestamp)
      ```

      #v(6pt)
      - FK constraints with `PRAGMA foreign_keys`
      - Indexes on `(service_id, metric_type, timestamp)`
      - Upsert for idempotent re-runs
    ]
  )
]

#content-slide([REST API])[
  #set text(size: 16pt)

  #table(
    columns: (1fr, 1fr),
    stroke: 0.5pt + border,
    inset: 8pt,
    fill: (x, y) => if y == 0 { raised } else { none },
    [*Endpoint*], [*Returns*],
    [`GET /api/healthz`], [Health check],
    [`GET /api/services`], [List all analyzed services],
    [`GET /api/overview`], [Repo aggregates + per-service summaries],
    [`GET /api/services/{id}/scores`], [OCM score time series],
    [`GET /api/services/{id}/metrics/{type}`], [Metric time series],
    [`GET /api/services/{id}/metrics/{type}/evidence`], [Evidence items for latest run],
  )

  #v(8pt)
  - Built with Go 1.22 `http.ServeMux` (method + path pattern matching)
  - All responses are JSON with proper `Content-Type`
  - CORS enabled for local development
]

// ─── 10. Dashboard ───
#section-slide([Dashboard])

#content-slide([Dashboard Design])[
  #set text(size: 16pt)

  Stripe-inspired dark theme with responsive layout:

  #v(8pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      *Repo Overview:*
      - Composite OCM average across all services
      - 6 metric tiles with aggregated values
      - Service count and API status

      *Service Overview:*
      - Individual service metrics
      - OCM score trend chart (Canvas)
      - Radar chart (normalized profile)
      - Bar chart (raw values)
      - Sparklines in metric tiles
    ],
    [
      *Interactive features:*
      - Sidebar service selector
      - "All Services" ↔ single-service toggle
      - Evidence drill-down modal per metric
      - Service comparison table
      - Responsive (mobile, tablet, desktop)

      *Tech:*
      - Pure HTML/CSS/JS (no frameworks)
      - Canvas-based charts
      - Embedded via `go:embed`
    ]
  )
]

// ─── 11. Evidence Traceability ───
#content-slide([Evidence-Based Traceability])[
  #set text(size: 16pt)

  Every metric contribution is tracked with provenance:

  #v(8pt)

  #table(
    columns: (auto, 1fr, 1fr),
    stroke: 0.5pt + border,
    inset: 8pt,
    fill: (x, y) => if y == 0 { raised } else { none },
    [*Metric*], [*Evidence Type*], [*Example*],
    [#metric-pill("CSA", csa-c)], [`env`, `port`, `resource`, `replica`], [`PORT=8080` from `api/deploy.yaml`],
    [#metric-pill("DD", dd-c)], [`dep_chain`, `direct_dep`], [`api → cache → db` (depth=2)],
    [#metric-pill("DB", db-c)], [`upstream_dep`, `downstream_dep`], [`orders → payment` edge],
    [#metric-pill("CV", cv-c)], [`commit`], [`commit abc12345`],
    [#metric-pill("FE", fe-c)], [`loadbalancer`, `ingress`, `external_url`], [LoadBalancer port 80],
    [#metric-pill("CDR", cdr-c)], [`env_override`], [`config.yaml` in 2 envs: dev, prod],
  )

  #v(8pt)
  Users can click any metric tile → modal shows exactly *what* contributed to the score and *where* it was found.
]

// ─── 12. Testing ───
#section-slide([Testing & Validation])

#content-slide([Test Suite])[
  #set text(size: 17pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      *Pipeline tests:*
      - DD cycle handling (Tarjan's SCC)
      - DB in/out degree counting
      - Self-dependency ignored
      - Normalization edge cases (max = min)
      - All 6 metrics present in output
      - Composite score in \[0, 1\]
      - DD evidence generation
    ],
    [
      *Parser tests:*
      - 7 dependency extraction tests
      - 6 FE detection tests
      - Suffix, FQDN, hostname:port patterns
      - Multi-document YAML

      *Storage tests:*
      - Round-trip for all 6 metric types
      - Empty source path handling

      *Git log tests:*
      - 6 tests for log parsing
    ]
  )

  #v(8pt)
  #align(center)[
    #box(fill: cdr-c.lighten(90%), radius: 6pt, inset: 10pt)[
      #text(fill: cdr-c, weight: 600, size: 16pt)[All 29 tests pass · Build clean · Vet clean]
    ]
  ]
]

// ─── 13. Project Structure ───
#content-slide([Project Structure])[
  #set text(size: 15pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 16pt,
    [
      ```
      ocm/
      ├── cmd/ocm/main.go        # CLI entry
      ├── internal/
      │   ├── model/model.go     # Types
      │   ├── parser/
      │   │   ├── yamlscan.go    # K8s parser
      │   │   └── parser_test.go
      │   ├── pipeline/
      │   │   ├── pipeline.go    # Orchestration
      │   │   └── pipeline_test.go
      │   ├── gitlog/
      │   │   ├── gitlog.go      # CV metric
      │   │   └── gitlog_test.go
      ```
    ],
    [
      ```
      │   ├── storage/
      │   │   ├── storage.go     # SQLite layer
      │   │   └── storage_test.go
      │   ├── api/api.go         # HTTP API
      │   └── dashboard/
      │       ├── dashboard.go   # go:embed
      │       └── static/
      │           ├── index.html
      │           ├── styles.css
      │           └── app.js
      ├── specs/                  # 8 spec docs
      ├── docs/                   # Phase docs
      └── go.mod                  # 2 deps
      ```
    ]
  )
]

// ─── 14. Phase Evolution ───
#section-slide([Project Evolution])

#content-slide([Phase 2 → Phase 3])[
  #set text(size: 17pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      *Phase 2 (MVP/PoC):*
      - 2 metrics: CSA + DD only
      - Basic dependency extraction
      - Simple dashboard
      - CSA evidence only
      - Composite: 50/50 weights
      - Manual validation
    ],
    [
      *Phase 3 (Current):*
      - All 6 metrics: CSA, DD, DB, CV, FE, CDR
      - Broadened dependency detection (9 patterns)
      - Git integration for CV
      - FE + CDR detection rules
      - Evidence for all 6 metrics
      - Stripe-inspired dashboard redesign
      - Repo-wide + service-specific views
      - Radar, bar, trend, sparkline charts
      - 29 automated tests
      - `--cv-window` CLI flag
      - Composite: equal 1/6 weights
    ]
  )
]

// ─── 15. Demo ───
#section-slide([Demo & Results])

#content-slide([Running OCM])[
  #set text(size: 17pt)

  *Analyze a Kubernetes repository:*

  #v(8pt)
  ```bash
  # Build and run
  go run ./cmd/ocm \
    --repo /path/to/microservices-demo \
    --db ocm.sqlite \
    --port 8080

  # Output
  ocm: analyzed 12 services, db=ocm.sqlite
  ocm: listening on http://127.0.0.1:8080
  ```

  #v(8pt)
  *JSON output mode (no server):*

  ```bash
  go run ./cmd/ocm --repo ./repo --print 2>/dev/null | jq '.services[].name'
  ```

  #v(8pt)
  *Configure analysis:*

  ```bash
  --service-key manifest    # Use metadata.name instead of directory
  --cv-window 60            # 60-day lookback for Change Volatility
  ```
]

// ─── 16. Known Limitations ───
#content-slide([Known Limitations])[
  #set text(size: 17pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      *By design (prototype scope):*
      - Heuristic-based extraction
      - No Helm template rendering
      - No Dockerfile parsing
      - Single-run-on-startup model
      - `git` binary required for CV
      - Localhost only (no auth)
    ],
    [
      *Mitigation strategies:*
      - Evidence trail enables manual verification
      - Multiple service ID strategies
      - Best-effort approach (warnings, not failures)
      - Modular architecture enables extension
      - Configurable weights for composite scoring
    ]
  )
]

// ─── 17. Future Work ───
#section-slide([Future Work])

#content-slide([Potential Extensions])[
  #set text(size: 17pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      *Near-term:*
      - Helm v3 template rendering
      - Dockerfile complexity parsing
      - Scheduled/on-demand re-analysis
      - Configurable weights via UI
      - Export reports (PDF/CSV)
    ],
    [
      *Long-term:*
      - CI/CD integration (PR complexity diffs)
      - Multi-repo support
      - Complexity budget enforcement
      - Runtime metrics correlation
      - Team/org-level dashboards
      - Authentication & multi-user
    ]
  )
]

// ─── 18. Summary ───
#content-slide([Summary])[
  #set text(size: 18pt)

  #grid(
    columns: (auto, 1fr),
    column-gutter: 12pt,
    row-gutter: 12pt,
    text(fill: accent, size: 20pt)[✓], [*6 metrics* quantifying operational complexity from K8s YAML],
    text(fill: accent, size: 20pt)[✓], [*Single binary* — Go + embedded SQLite + embedded dashboard],
    text(fill: accent, size: 20pt)[✓], [*Evidence-based* — every metric contribution is traceable],
    text(fill: accent, size: 20pt)[✓], [*Graph algorithms* — Tarjan's SCC for cycle-safe dependency depth],
    text(fill: accent, size: 20pt)[✓], [*Minimal dependencies* — only `yaml.v3` + pure-Go SQLite],
    text(fill: accent, size: 20pt)[✓], [*Interactive dashboard* — repo overview, service drill-down, charts],
    text(fill: accent, size: 20pt)[✓], [*29 automated tests* — parser, pipeline, storage, git log],
  )
]

// ─── 19. End ───
#slide[
  #set page(fill: bg)
  #set text(fill: txt)
  #set align(horizon + center)
  #block(width: 100%)[
    #text(size: 38pt, weight: 700)[Thank You]
    #v(12pt)
    #text(size: 18pt, fill: txt2)[Questions?]
    #v(20pt)
    #line(length: 80pt, stroke: 1.5pt + accent)
    #v(16pt)
    #text(size: 14pt, fill: txt3)[
      Aabid Ali Sofi #h(8pt) | #h(8pt) 2023EBCS041 \
      Operational Complexity Meter — Phase 3
    ]
  ]
]
