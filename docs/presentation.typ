// OCM — Operational Complexity Meter
// Full Project Presentation (Polylux)

#import "@preview/polylux:0.4.0": *

#set page(paper: "presentation-16-9")
#set text(size: 20pt, font: ("Inter", "Inter Variable"))

// ─── Color palette ───
#let bg      = rgb("#0c1017")
#let surface = rgb("#12161e")
#let raised  = rgb("#1a1f2b")
#let hover   = rgb("#222836")
#let border  = rgb("#2d3343")
#let txt     = rgb("#eef0f4")
#let txt2    = rgb("#8b93a1")
#let txt3    = rgb("#5f6878")
#let accent  = rgb("#635bff")
#let accent2 = rgb("#818cf8")
#let csa-c   = rgb("#635bff")
#let dd-c    = rgb("#0ea5e9")
#let db-c    = rgb("#8b5cf6")
#let cv-c    = rgb("#f59e0b")
#let fe-c    = rgb("#ef4444")
#let cdr-c   = rgb("#10b981")
#let success = rgb("#10b981")

// ─── Reusable helpers ───
#let slide-footer() = place(bottom + right, dx: -24pt, dy: -12pt,
  text(size: 10pt, fill: txt3)[OCM — Phase 3]
)

#let title-slide(title, subtitle) = slide[
  #set page(fill: bg)
  #set text(fill: txt)
  #set align(horizon + center)
  #block(width: 100%)[
    #box(fill: accent.lighten(88%), radius: 20pt, inset: (x: 14pt, y: 6pt))[
      #text(size: 12pt, fill: accent, weight: 700, tracking: 0.15em)[STUDY PROJECT — PHASE 3]
    ]
    #v(18pt)
    #text(size: 48pt, weight: 800, tracking: -0.02em)[#title]
    #v(6pt)
    #text(size: 21pt, fill: txt2, weight: 400)[#subtitle]
    #v(28pt)
    #line(length: 100pt, stroke: 2pt + accent)
    #v(18pt)
    #text(size: 14pt, fill: txt3, weight: 500)[
      Aabid Ali Sofi #h(16pt) | #h(16pt) 2023EBCS041 \
      Advisor: Preethy P Johny #h(16pt) | #h(16pt) Feb 2026
    ]
  ]
]

#let section-slide(title) = slide[
  #set page(fill: surface)
  #set text(fill: txt)
  #set align(horizon + center)
  #block(width: 100%)[
    #box(fill: accent.lighten(88%), radius: 16pt, inset: (x: 12pt, y: 5pt))[
      #text(size: 11pt, fill: accent, weight: 700, tracking: 0.12em)[SECTION]
    ]
    #v(10pt)
    #text(size: 42pt, weight: 800, tracking: -0.02em)[#title]
    #v(14pt)
    #line(length: 70pt, stroke: 2pt + accent)
  ]
  #slide-footer()
]

#let content-slide(title, body) = slide[
  #set page(fill: bg)
  #set text(fill: txt)
  #v(4pt)
  #text(size: 30pt, weight: 700, tracking: -0.01em)[#title]
  #v(4pt)
  #line(length: 100%, stroke: 0.5pt + border)
  #v(10pt)
  #text(size: 17pt, fill: txt2)[#body]
  #slide-footer()
]

#let metric-pill(label, color) = box(
  fill: color.lighten(85%),
  radius: 6pt,
  inset: (x: 10pt, y: 4pt),
  text(size: 13pt, weight: 700, fill: color)[#label]
)

#let callout(body, color: accent) = align(center)[
  #block(
    fill: color.lighten(92%),
    radius: 10pt,
    inset: (x: 20pt, y: 12pt),
    width: auto,
    stroke: 0.5pt + color.lighten(70%),
  )[
    #text(fill: color, weight: 600, size: 16pt)[#body]
  ]
]

#let step-list(color: accent, items) = {
  grid(
    columns: (auto, 1fr),
    column-gutter: 14pt,
    row-gutter: 12pt,
    ..items.enumerate().map(((i, item)) => (
      box(
        fill: color.lighten(85%),
        radius: 14pt,
        width: 28pt, height: 28pt,
        align(center + horizon, text(fill: color, weight: 800, size: 14pt)[#{i + 1}])
      ),
      align(horizon, item),
    )).flatten()
  )
}

#let screenshot-frame(path, caption: none) = {
  align(center)[
    #block(
      radius: 10pt,
      clip: true,
      stroke: 0.5pt + border,
    )[
      #image(path)
    ]
    #if caption != none {
      v(6pt)
      text(size: 12pt, fill: txt3, style: "italic")[#caption]
    }
  ]
}


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
    column-gutter: 32pt,
    [
      #step-list(color: accent, (
        [Problem Statement],
        [What is OCM?],
        [The 6 Metrics],
        [System Architecture],
        [Pipeline Deep Dive],
      ))
    ],
    [
      #step-list(color: accent2, (
        [Parser & Extraction],
        [Graph Algorithms],
        [Storage & API],
        [Dashboard & Visualization],
        [Demo, Results & Future Work],
      ))
    ]
  )
]

// ─── 3. Problem Statement ───
#section-slide([Problem Statement])

#content-slide([The Problem])[
  Modern distributed systems are *operationally complex*, but that complexity is
  rarely measured.

  #v(10pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 24pt,
    [
      #block(fill: raised, radius: 10pt, inset: 14pt, width: 100%)[
        #text(fill: fe-c, weight: 700, size: 15pt)[Symptoms]
        #v(6pt)
        #set text(size: 16pt, fill: txt2)
        - Hundreds of config parameters
        - Deep dependency chains
        - Frequent config changes
        - Env-specific drift (dev vs prod)
        - Exposed failure surfaces
      ]
    ],
    [
      #block(fill: raised, radius: 10pt, inset: 14pt, width: 100%)[
        #text(fill: cv-c, weight: 700, size: 15pt)[Impact]
        #v(6pt)
        #set text(size: 16pt, fill: txt2)
        - Outages from config errors
        - Slow incident response
        - Hidden coupling between services
        - Unmeasured toil & risk
      ]
    ]
  )

  #v(12pt)
  #callout[Can we quantify operational complexity as a single, measurable score?]
]

// ─── 4. What is OCM? ───
#section-slide([What is OCM?])

#content-slide([Operational Complexity Meter])[
  #set text(size: 17pt)

  OCM is a *single Go binary* that:

  #v(8pt)

  #step-list(color: accent, (
    [*Scans* a Kubernetes YAML repository],
    [*Identifies* services from directory structure or manifest metadata],
    [*Computes* 6 operational complexity metrics per service],
    [*Normalizes* and produces a composite OCM score (0--1)],
    [*Stores* results in SQLite for time-series tracking],
    [*Serves* an interactive web dashboard],
  ))

  #v(12pt)

  #block(fill: raised, radius: 8pt, inset: 12pt, width: 100%)[
    #set text(size: 15pt, fill: txt2)
    ```bash
    go run ./cmd/ocm --repo /path/to/k8s-repo --db ocm.sqlite --port 8080
    ```
  ]

  #v(4pt)
  #text(size: 13pt, fill: txt3, weight: 500)[
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
    inset: 10pt,
    fill: (x, y) => if y == 0 { raised } else if calc.odd(y) { surface } else { none },

    [*Metric*], [*What It Measures*], [*How*],

    [#metric-pill("CSA", csa-c)], [Configuration Surface Area], [Count of env vars, ports, resources, replicas, spec keys],
    [#metric-pill("DD", dd-c)], [Dependency Depth], [Longest path in the SCC-condensed dependency DAG],
    [#metric-pill("DB", db-c)], [Dependency Breadth], [Direct upstream + downstream dependency count],
    [#metric-pill("CV", cv-c)], [Change Volatility], [Git commits affecting config in the last 30 days],
    [#metric-pill("FE", fe-c)], [Failure Exposure], [Exposed endpoints (LB, NodePort, Ingress, hostPort, ext URLs)],
    [#metric-pill("CDR", cdr-c)], [Config Drift Risk], [Env-specific overrides (dev/staging/prod directory diffs)],
  )
]

#content-slide([Metric Formulas])[
  #set text(size: 15pt)
  #grid(
    columns: (1fr),
    column-gutter: 16pt,
    row-gutter: 12pt,
    [
      #block(fill: raised, radius: 8pt, inset: 12pt, width: 100%)[
        #metric-pill("CSA", csa-c)
        #v(6pt)
        $ "CSA"(s) = |"env"| + |"ports"| + |"resources"| + |"replicas"| + |"spec keys"| $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 12pt, width: 100%)[
        #metric-pill("DD", dd-c)
        #v(6pt)
        $ "DD"(s) = max_(p in "paths from" s) |p| quad "(edges in SCC-condensed DAG)" $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 12pt, width: 100%)[
        #metric-pill("DB", db-c)
        #v(6pt)
        $ "DB"(s) = "deg"^+(s) + "deg"^-(s) $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 12pt, width: 100%)[
        #metric-pill("CV", cv-c)
        #v(6pt)
        $ "CV"(s) = |{ c in "commits" : "touches"(c, s) and c in [t - Delta, t] }| $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 12pt, width: 100%)[
        #metric-pill("FE", fe-c)
        #v(6pt)
        $ "FE"(s) = |"LB ports"| + |"NP ports"| + |"Ingress paths"| + |"hostPorts"| + |"ext URLs"| $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 12pt, width: 100%)[
        #metric-pill("CDR", cdr-c)
        #v(6pt)
        $ "CDR"(s) = sum_(f in "config files") (|"envs"(f)| - 1)^+ $
      ]
    ],
  )

  #v(6pt)
  #text(size: 12pt, fill: txt3)[
    $s$ = service, $"deg"^+$ = out-degree, $"deg"^-$ = in-degree, $Delta$ = lookback window (default 30 d), $(dot)^+$ = $max(dot, 0)$
  ]
]

#content-slide([Composite OCM Score])[
  #set text(size: 18pt)

  #block(fill: raised, radius: 10pt, inset: 16pt, width: 100%)[
    #text(fill: accent2, weight: 700, size: 15pt)[Normalization]
    #text(fill: txt2, size: 16pt)[ — Min-max across the service cohort:]
    #v(8pt)
    #align(center)[
      $ "Normalized"(M) = (M - M_"min") / (M_"max" - M_"min") $
    ]
  ]

  #v(14pt)

  #block(fill: raised, radius: 10pt, inset: 16pt, width: 100%)[
    #text(fill: accent, weight: 700, size: 15pt)[Composite Score]
    #text(fill: txt2, size: 16pt)[ — Weighted sum of normalized metrics:]
    #v(8pt)
    #align(center)[
      $ "OCM" = 1/6 dot "CSA" + 1/6 dot "DD" + 1/6 dot "DB" + 1/6 dot "CV" + 1/6 dot "FE" + 1/6 dot "CDR" $
    ]
  ]

  #v(14pt)

  #grid(
    columns: (1fr, 1fr, 1fr),
    column-gutter: 16pt,
    align(center)[
      #text(fill: success, weight: 700, size: 24pt)[0.0]
      #v(2pt)
      #text(fill: txt3, size: 13pt)[Minimal complexity]
    ],
    align(center)[
      #text(fill: txt3, size: 13pt)[Weights are *configurable*\ (default: equal 1/6 each)]
    ],
    align(center)[
      #text(fill: fe-c, weight: 700, size: 24pt)[1.0]
      #v(2pt)
      #text(fill: txt3, size: 13pt)[Maximum in cohort]
    ],
  )
]

// ─── 6. Architecture ───
#section-slide([System Architecture])

#content-slide([Single-Binary Architecture])[
  #set text(size: 16pt)
  #align(center)[
    #block(fill: raised, radius: 10pt, inset: 16pt, width: 92%)[
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

  #v(10pt)
  #grid(
    columns: (1fr, 1fr, 1fr),
    column-gutter: 12pt,
    block(fill: surface, radius: 8pt, inset: 12pt, width: 100%)[
      #text(fill: accent, weight: 700, size: 14pt)[2 Dependencies]
      #v(4pt)
      #text(fill: txt2, size: 14pt)[
        `yaml.v3` \
        `modernc.org/sqlite`
      ]
    ],
    block(fill: surface, radius: 8pt, inset: 12pt, width: 100%)[
      #text(fill: dd-c, weight: 700, size: 14pt)[No CGO]
      #v(4pt)
      #text(fill: txt2, size: 14pt)[
        Pure Go SQLite \
        Cross-compilable
      ]
    ],
    block(fill: surface, radius: 8pt, inset: 12pt, width: 100%)[
      #text(fill: cdr-c, weight: 700, size: 14pt)[Embedded Assets]
      #v(4pt)
      #text(fill: txt2, size: 14pt)[
        `go:embed` HTML/CSS/JS \
        Zero external files
      ]
    ],
  )
]

#content-slide([Data Flow Pipeline])[
  #set text(size: 17pt)
  #step-list(color: accent, (
    [*Discovery* — Walk filesystem, collect `.yaml`/`.yml`, skip `.git`/`vendor`],
    [*Parsing* — Multi-doc YAML decode, extract CSA facts + deps + FE signals],
    [*Service ID* — Map files to services via `dir` or `manifest` (metadata.name)],
    [*Metric Engine* — Compute DD (graph), DB (degree), CV (git), FE, CDR],
    [*Normalization* — Min-max across cohort for each metric],
    [*Composite* — Weighted sum to single OCM score per service],
    [*Persist* — Atomic transaction: services + metrics + evidence to SQLite],
    [*Serve* — HTTP API + embedded dashboard at `localhost:8080`],
  ))
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
    inset: 10pt,
    fill: (x, y) => if y == 0 { raised } else if calc.odd(y) { surface } else { none },
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

  #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
    #text(fill: dd-c, weight: 700, size: 14pt)[Rule 1 — Env var suffix matching]
    #v(4pt)
    #text(size: 14pt, fill: txt2)[
      `_SERVICE`, `_SERVICE_HOST`, `_SERVICE_PORT`, `_ADDR`, `_HOST`, `_URL`, `_ENDPOINT`, `_URI`
    ]
  ]

  #v(8pt)

  #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
    #text(fill: db-c, weight: 700, size: 14pt)[Rule 2 — Cluster FQDN detection]
    #v(4pt)
    #text(size: 14pt, fill: txt2)[
      Values containing `.svc.cluster.local` — extract service name
    ]
  ]

  #v(8pt)

  #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
    #text(fill: cdr-c, weight: 700, size: 14pt)[Rule 3 — Hostname:port pattern]
    #v(4pt)
    #text(size: 14pt, fill: txt2)[
      Regex: `^[a-z][a-z0-9-]{0,62}:\d{2,5}$` — extract hostname
    ]
  ]

  #v(8pt)
  #text(size: 13pt, fill: txt3, weight: 500)[
    Designed for real-world repos: Sock Shop, Google Online Boutique, Helm charts, Rails/Django/Spring deployments.
  ]
]

#content-slide([Failure Exposure Detection])[
  #set text(size: 16pt)

  FE counts exposed endpoints and external integrations — 5 detection rules:

  #v(8pt)

  #step-list(color: fe-c, (
    [*LoadBalancer / NodePort* service types — each port = 1 endpoint],
    [*Ingress* resources — each `rules[].http.paths[]` = 1 endpoint],
    [*ExternalName* services — 1 external integration],
    [*Container hostPort* — each hostPort = 1 endpoint],
    [*External URLs* in env vars — `http(s)://` not containing `.svc.cluster.local`],
  ))

  #v(10pt)
  #text(size: 13pt, fill: txt3, weight: 500)[
    Detection rules are explicitly documented in code per spec requirement.
  ]
]

// ─── 8. Graph Algorithms ───
#section-slide([Graph Algorithms])

#content-slide([Dependency Depth — Tarjan's SCC])[
  #set text(size: 16pt)

  DD = longest path (edge count) in the dependency graph.

  #v(8pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: fe-c, weight: 700, size: 14pt)[Problem]
        #v(4pt)
        #text(fill: txt2, size: 15pt)[Dependency graphs may have cycles (A -> B -> A).]
      ]

      #v(10pt)

      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 14pt)[Solution — SCC Condensation]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        #step-list(color: dd-c, (
          [Tarjan's algorithm finds SCCs],
          [Condense each SCC to a single node],
          [Result is a DAG (no cycles)],
          [DFS with memoization -> longest path],
        ))
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: dd-c, weight: 700, size: 14pt)[Example]
        #v(6pt)
        #text(fill: txt2, size: 14pt)[
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
      ]

      #v(8pt)
      #text(size: 13pt, fill: txt3, weight: 500)[
        Deterministic — same input always produces same DD values.
      ]
    ]
  )
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
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 14pt)[Core Tables]
        #v(6pt)
        #text(fill: txt2, size: 14pt)[
          ```sql
          services(id, name, repository)
            UNIQUE(name, repository)

          metrics(id, service_id,
            metric_type, metric_value,
            timestamp)

          composite_scores(id, service_id,
            ocm_score, timestamp)
          ```
        ]
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 14pt)[Evidence Table]
        #v(6pt)
        #text(fill: txt2, size: 14pt)[
          ```sql
          metric_evidence(id, service_id,
            metric_type, component,
            evidence_key, evidence_value,
            source_path,
            manifest_kind, manifest_name,
            timestamp)
          ```
        ]
      ]

      #v(6pt)
      #set text(size: 14pt, fill: txt2)
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
    inset: 10pt,
    fill: (x, y) => if y == 0 { raised } else if calc.odd(y) { surface } else { none },
    [*Endpoint*], [*Returns*],
    [`GET /api/healthz`], [Health check],
    [`GET /api/services`], [List all analyzed services],
    [`GET /api/overview`], [Repo aggregates + per-service summaries],
    [`GET /api/services/{id}/scores`], [OCM score time series],
    [`GET /api/services/{id}/metrics/{type}`], [Metric time series],
    [`GET /api/services/{id}/metrics/{type}/evidence`], [Evidence items for latest run],
  )

  #v(10pt)
  #grid(
    columns: (1fr, 1fr, 1fr),
    column-gutter: 12pt,
    block(fill: surface, radius: 8pt, inset: 10pt, width: 100%)[
      #text(fill: txt3, size: 13pt)[Go 1.22 `http.ServeMux` \ method + path patterns]
    ],
    block(fill: surface, radius: 8pt, inset: 10pt, width: 100%)[
      #text(fill: txt3, size: 13pt)[JSON responses \ proper `Content-Type`]
    ],
    block(fill: surface, radius: 8pt, inset: 10pt, width: 100%)[
      #text(fill: txt3, size: 13pt)[CORS enabled \ for local development]
    ],
  )
]

// ─── 10. Dashboard ───
#section-slide([Dashboard])

#content-slide([Dashboard — Repository Overview])[
  #set text(size: 15pt)

  #screenshot-frame("screenshots/dashboard-overview.png",
    caption: [Repository overview showing composite OCM score, 9 analyzed services, and the 6 metric breakdown tiles]
  )
]

#content-slide([Dashboard — Service Detail])[
  #set text(size: 15pt)

  #screenshot-frame("screenshots/dashboard-service-tall.png",
    caption: [Single service view: individual metrics, OCM score trend, radar chart, and bar chart for the `api` service]
  )
]

#content-slide([Dashboard — Service Comparison])[
  #set text(size: 15pt)

  #screenshot-frame("screenshots/dashboard-services-table.png",
    caption: [All 9 services compared side-by-side with raw metric values and composite OCM scores]
  )
]

#content-slide([Dashboard — Evidence Drill-Down])[
  #set text(size: 15pt)

  #screenshot-frame("screenshots/dashboard-evidence.png",
    caption: [CSA evidence modal for the `api` service — 25 configuration contributors with source traceability]
  )
]

#content-slide([Dashboard Design])[
  #set text(size: 16pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 24pt,
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 14pt)[Views & Visualizations]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - Composite OCM score hero card
        - 6 metric tiles with color coding
        - OCM score trend chart (Canvas)
        - Radar chart (normalized profile)
        - Bar chart (raw metric values)
        - Service comparison table
        - Evidence drill-down modal
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 14pt)[Technical Implementation]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - Stripe-inspired dark design system
        - Pure HTML/CSS/JS (no frameworks)
        - Canvas-based chart rendering
        - Responsive layout (mobile to desktop)
        - Embedded via `go:embed`
        - Sidebar with live service selector
        - All Services vs individual service toggle
      ]
    ]
  )
]

// ─── 11. Evidence Traceability ───
#content-slide([Evidence-Based Traceability])[
  #set text(size: 15pt)

  Every metric contribution is tracked with provenance:

  #v(6pt)

  #table(
    columns: (auto, 1fr, 1fr),
    stroke: 0.5pt + border,
    inset: 9pt,
    fill: (x, y) => if y == 0 { raised } else if calc.odd(y) { surface } else { none },
    [*Metric*], [*Evidence Type*], [*Example*],
    [#metric-pill("CSA", csa-c)], [`env`, `port`, `resource`, `replica`], [`PORT=8080` from `api/deploy.yaml`],
    [#metric-pill("DD", dd-c)], [`dep_chain`, `direct_dep`], [`api -> cache -> db` (depth=2)],
    [#metric-pill("DB", db-c)], [`upstream_dep`, `downstream_dep`], [`orders -> payment` edge],
    [#metric-pill("CV", cv-c)], [`commit`], [`commit abc12345`],
    [#metric-pill("FE", fe-c)], [`loadbalancer`, `ingress`, `external_url`], [LoadBalancer port 80],
    [#metric-pill("CDR", cdr-c)], [`env_override`], [`config.yaml` in 2 envs: dev, prod],
  )

  #v(6pt)
  #callout(color: accent2)[Click any metric tile in the dashboard to see exactly *what* contributed and *where* it was found.]
]

// ─── 12. Testing ───
#section-slide([Testing & Validation])

#content-slide([Test Suite])[
  #set text(size: 16pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 14pt)[Pipeline Tests]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - DD cycle handling (Tarjan's SCC)
        - DB in/out degree counting
        - Self-dependency ignored
        - Normalization edge cases (max = min)
        - All 6 metrics present in output
        - Composite score in \[0, 1\]
        - DD evidence generation
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: dd-c, weight: 700, size: 14pt)[Parser Tests]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - 7 dependency extraction tests
        - 6 FE detection tests
        - Suffix, FQDN, hostname:port patterns
        - Multi-document YAML
      ]

      #v(8pt)

      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 14pt)[Storage & Git Tests]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - Round-trip for all 6 metric types
        - 6 git log parsing tests
      ]
    ]
  )

  #v(8pt)
  #callout(color: success)[All 29 tests pass -- Build clean -- Vet clean]
]

// ─── 13. Project Structure ───
#content-slide([Project Structure])[
  #set text(size: 14pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 16pt,
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: txt2)[
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
        ]
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: txt2)[
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
      ]
    ]
  )
]

// ─── 14. Phase Evolution ───
#section-slide([Project Evolution])

#content-slide([Phase 2 -> Phase 3])[
  #set text(size: 16pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: txt3, weight: 700, size: 15pt)[Phase 2 — MVP/PoC]
        #v(6pt)
        #set text(size: 15pt, fill: txt3)
        - 2 metrics: CSA + DD only
        - Basic dependency extraction
        - Simple dashboard
        - CSA evidence only
        - Composite: 50/50 weights
        - Manual validation
      ]
    ],
    [
      #block(fill: accent.lighten(92%), radius: 8pt, inset: 14pt, width: 100%, stroke: 0.5pt + accent.lighten(70%))[
        #text(fill: accent, weight: 700, size: 15pt)[Phase 3 — Current]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - All 6 metrics: CSA, DD, DB, CV, FE, CDR
        - Broadened dependency detection (9 patterns)
        - Git integration for CV
        - FE + CDR detection rules
        - Evidence for all 6 metrics
        - Stripe-inspired dashboard redesign
        - Radar, bar, trend, sparkline charts
        - 29 automated tests
        - `--cv-window` CLI flag
      ]
    ]
  )
]

// ─── 15. Demo ───
#section-slide([Demo & Results])

#content-slide([Running OCM])[
  #set text(size: 16pt)

  #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
    #text(fill: accent, weight: 700, size: 14pt)[Analyze a Kubernetes repository]
    #v(6pt)
    #text(fill: txt2, size: 15pt)[
      ```bash
      # Build and run
      go run ./cmd/ocm \
        --repo /path/to/microservices-demo \
        --db ocm.sqlite --port 8080

      # Output
      ocm: analyzed 12 services, db=ocm.sqlite
      ocm: listening on http://127.0.0.1:8080
      ```
    ]
  ]

  #v(10pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 16pt,
    block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
      #text(fill: cv-c, weight: 700, size: 14pt)[JSON output mode]
      #v(4pt)
      #text(fill: txt2, size: 14pt)[
        ```bash
        go run ./cmd/ocm --repo ./repo \
          --print 2>/dev/null | jq '.'
        ```
      ]
    ],
    block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
      #text(fill: cdr-c, weight: 700, size: 14pt)[Configuration flags]
      #v(4pt)
      #text(fill: txt2, size: 14pt)[
        ```bash
        --service-key manifest  # metadata.name
        --cv-window 60          # 60-day lookback
        ```
      ]
    ],
  )
]

// ─── 16. Known Limitations ───
#content-slide([Known Limitations])[
  #set text(size: 16pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: cv-c, weight: 700, size: 14pt)[By Design (Study project)]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - Heuristic-based extraction
        - No Helm template rendering
        - No Dockerfile parsing
        - Single-run-on-startup model
        - `git` binary required for CV
        - Localhost only (no auth)
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 14pt)[Mitigations]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - Evidence trail enables manual verification
        - Multiple service ID strategies
        - Best-effort approach (warnings, not failures)
        - Modular architecture enables extension
        - Configurable weights for composite scoring
      ]
    ]
  )
]

// ─── 17. Future Work ───
#section-slide([Future Work])

#content-slide([Potential Extensions])[
  #set text(size: 16pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 20pt,
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 14pt)[Near-Term]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - Helm v3 template rendering
        - Dockerfile complexity parsing
        - Scheduled/on-demand re-analysis
        - Configurable weights via UI
        - Export reports (PDF/CSV)
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 14pt, width: 100%)[
        #text(fill: accent2, weight: 700, size: 14pt)[Long-Term]
        #v(6pt)
        #set text(size: 15pt, fill: txt2)
        - CI/CD integration (PR complexity diffs)
        - Multi-repo support
        - Complexity budget enforcement
        - Runtime metrics correlation
        - Team/org-level dashboards
        - Authentication & multi-user
      ]
    ]
  )
]

// ─── 18. Summary ───
#content-slide([Summary])[
  #set text(size: 17pt)

  #grid(
    columns: (auto, 1fr),
    column-gutter: 14pt,
    row-gutter: 14pt,
    box(fill: success.lighten(85%), radius: 14pt, width: 28pt, height: 28pt,
      align(center + horizon, text(fill: success, weight: 800, size: 16pt)[+])
    ), align(horizon)[*6 metrics* quantifying operational complexity from K8s YAML],

    box(fill: success.lighten(85%), radius: 14pt, width: 28pt, height: 28pt,
      align(center + horizon, text(fill: success, weight: 800, size: 16pt)[+])
    ), align(horizon)[*Single binary* — Go + embedded SQLite + embedded dashboard],

    box(fill: success.lighten(85%), radius: 14pt, width: 28pt, height: 28pt,
      align(center + horizon, text(fill: success, weight: 800, size: 16pt)[+])
    ), align(horizon)[*Evidence-based* — every metric contribution is traceable],

    box(fill: success.lighten(85%), radius: 14pt, width: 28pt, height: 28pt,
      align(center + horizon, text(fill: success, weight: 800, size: 16pt)[+])
    ), align(horizon)[*Graph algorithms* — Tarjan's SCC for cycle-safe dependency depth],

    box(fill: success.lighten(85%), radius: 14pt, width: 28pt, height: 28pt,
      align(center + horizon, text(fill: success, weight: 800, size: 16pt)[+])
    ), align(horizon)[*Minimal dependencies* — only `yaml.v3` + pure-Go SQLite],

    box(fill: success.lighten(85%), radius: 14pt, width: 28pt, height: 28pt,
      align(center + horizon, text(fill: success, weight: 800, size: 16pt)[+])
    ), align(horizon)[*Interactive dashboard* — repo overview, service drill-down, charts],

    box(fill: success.lighten(85%), radius: 14pt, width: 28pt, height: 28pt,
      align(center + horizon, text(fill: success, weight: 800, size: 16pt)[+])
    ), align(horizon)[*29 automated tests* — parser, pipeline, storage, git log],
  )
]

// ─── 19. End ───
#slide[
  #set page(fill: bg)
  #set text(fill: txt)
  #set align(horizon + center)
  #block(width: 100%)[
    #text(size: 44pt, weight: 800)[Thank You]
    #v(12pt)
    #text(size: 20pt, fill: txt2)[Questions?]
    #v(24pt)
    #line(length: 80pt, stroke: 2pt + accent)
    #v(18pt)
    #text(size: 14pt, fill: txt3, weight: 500)[
      Aabid Ali Sofi #h(12pt) | #h(12pt) 2023EBCS041 \
      Operational Complexity Meter — Phase 3
    ]
  ]
]
