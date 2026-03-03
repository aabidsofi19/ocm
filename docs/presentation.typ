// OCM — Operational Complexity Meter
// Full Project Presentation (Polylux) — Consolidated (17 slides)

#import "@preview/polylux:0.4.0": *

#set page(paper: "presentation-16-9", margin: (x: 28pt, top: 20pt, bottom: 24pt))
#set text(size: 18pt, font: "Inter")

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

#let content-slide(title, body) = slide[
  #set page(fill: bg)
  #set text(fill: txt)
  #v(2pt)
  #text(size: 26pt, weight: 700, tracking: -0.01em)[#title]
  #v(1pt)
  #line(length: 100%, stroke: 0.5pt + border)
  #v(4pt)
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
    inset: (x: 20pt, y: 10pt),
    width: auto,
    stroke: 0.5pt + color.lighten(70%),
  )[
    #text(fill: color, weight: 600, size: 15pt)[#body]
  ]
]

#let step-list(color: accent, items) = {
  grid(
    columns: (auto, 1fr),
    column-gutter: 10pt,
    row-gutter: 8pt,
    ..items.enumerate().map(((i, item)) => (
      box(
        fill: color.lighten(85%),
        radius: 14pt,
        width: 24pt, height: 24pt,
        align(center + horizon, text(fill: color, weight: 800, size: 12pt)[#{i + 1}])
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
      v(4pt)
      text(size: 11pt, fill: txt3, style: "italic")[#caption]
    }
  ]
}


// ═══════════════════════════════════════════════════════
//  SLIDES (17 total)
// ═══════════════════════════════════════════════════════

// ─── 1. Title ───
#title-slide(
  [Operational Complexity Meter],
  [Quantifying Operational Complexity in Distributed Systems]
)

// ─── 2. Agenda ───
#content-slide([Agenda])[
  #set text(size: 17pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 28pt,
    [
      #step-list(color: accent, (
        [Problem Statement & OCM Overview],
        [The 6 Metrics & Composite Score],
        [System Architecture & Pipeline],
        [Parser, Graph Algorithms],
      ))
    ],
    [
      #step-list(color: accent2, (
        [Storage & REST API],
        [Dashboard & Visualization],
        [Testing & Validation],
        [Limitations, Future Work & Summary],
      ))
    ]
  )
]

// ─── 3. Problem Statement & OCM Overview ───
#content-slide([Problem Statement])[
  #set text(size: 13pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 16pt,
    [
      #block(fill: raised, radius: 8pt, inset: 10pt, width: 100%)[
        #text(fill: fe-c, weight: 700, size: 13pt)[Symptoms]
        #v(2pt)
        #set text(size: 13pt, fill: txt2)
        - Hundreds of config parameters
        - Deep dependency chains
        - Frequent config changes
        - Environment drift (dev vs prod)
        - Exposed failure surfaces
      ]
      #v(4pt)
      #block(fill: raised, radius: 8pt, inset: 10pt, width: 100%)[
        #text(fill: cv-c, weight: 700, size: 13pt)[Impact]
        #v(2pt)
        #set text(size: 13pt, fill: txt2)
        - Outages from config errors
        - Slow incident response
        - Hidden coupling between services
        - Unmeasured toil & risk
      ]
    ],
    [
      #block(fill: accent.lighten(92%), radius: 8pt, inset: 10pt, width: 100%, stroke: 0.5pt + accent.lighten(70%))[
        #text(fill: accent, weight: 700, size: 13pt)[OCM — The Solution]
        #v(2pt)
        #set text(size: 13pt, fill: txt2)
        A *single Go binary* that:
        #v(2pt)
        #step-list(color: accent, (
          [*Scans* a Kubernetes YAML repository],
          [*Computes* 6 complexity metrics per service],
          [*Produces* a composite OCM score (0--1)],
          [*Stores* results in SQLite],
          [*Serves* an interactive web dashboard],
        ))
      ]
      #v(4pt)
      #block(fill: raised, radius: 6pt, inset: 6pt, width: 100%)[
        #text(size: 11pt, fill: txt3)[```bash
        go run ./cmd/ocm --repo ./k8s-repo --db ocm.sqlite --port 8080
        ```]
      ]
    ]
  )
  #v(2pt)
  #callout[Can we quantify operational complexity as a single, measurable score?]
]

// ─── 4. The 6 Metrics — Overview ───
#content-slide([The 6 Metrics])[
  #set text(size: 14pt)
  #table(
    columns: (auto, 1fr, 1fr),
    align: (left, left, left),
    stroke: 0.5pt + border,
    inset: 6pt,
    fill: (x, y) => if y == 0 { raised } else if calc.odd(y) { surface } else { none },

    [*Metric*], [*What It Measures*], [*How*],
    [#metric-pill("CSA", csa-c)], [Configuration Surface Area], [Count of env vars, ports, resources, replicas, spec keys],
    [#metric-pill("DD", dd-c)], [Dependency Depth], [Longest path in SCC-condensed dependency DAG],
    [#metric-pill("DB", db-c)], [Dependency Breadth], [Direct upstream + downstream dependency count],
    [#metric-pill("CV", cv-c)], [Change Volatility], [Git commits affecting config in last 30 days],
    [#metric-pill("FE", fe-c)], [Failure Exposure], [Exposed endpoints (LB, NodePort, Ingress, hostPort, ext URLs)],
    [#metric-pill("CDR", cdr-c)], [Config Drift Risk], [Env-specific overrides (dev/staging/prod directory diffs)],
  )

  #v(4pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 16pt,
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: accent2, weight: 700, size: 12pt)[Normalization (Min-Max)]
        #v(2pt)
        #text(size: 14pt)[$ "Norm"(M) = (M - M_"min") / (M_"max" - M_"min") $]
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 12pt)[Composite OCM Score]
        #v(2pt)
        #text(size: 14pt)[$ "OCM" = 1/6 sum_(m in M) "Norm"(m) $]
      ]
    ],
  )
]

// ─── 5. Metric Formulas ───
#content-slide([Metric Formulas])[
  #set text(size: 13pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 12pt,
    row-gutter: 8pt,
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #metric-pill("CSA", csa-c)
        #h(4pt) $ "CSA"(s) = |"env"| + |"ports"| + |"resources"| + |"replicas"| + |"spec keys"| $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #metric-pill("DD", dd-c)
        #h(4pt) $ "DD"(s) = max_(p in "paths") |p| quad "(SCC DAG)" $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #metric-pill("DB", db-c)
        #h(4pt) $ "DB"(s) = "deg"^+(s) + "deg"^-(s) $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #metric-pill("CV", cv-c)
        #h(4pt) $ "CV"(s) = |{ c : "touches"(c, s) and c in [t - Delta, t] }| $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #metric-pill("FE", fe-c)
        #h(4pt) $ "FE"(s) = |"LB"| + |"NP"| + |"Ingress"| + |"hostPorts"| + |"ext URLs"| $
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #metric-pill("CDR", cdr-c)
        #h(4pt) $ "CDR"(s) = sum_(f) (|"envs"(f)| - 1)^+ $
      ]
    ],
  )

  #v(4pt)
  #grid(
    columns: (1fr, 1fr, 1fr),
    column-gutter: 12pt,
    align(center)[
      #text(fill: success, weight: 700, size: 20pt)[0.0]
      #text(fill: txt3, size: 11pt)[Minimal complexity]
    ],
    align(center)[
      #text(fill: txt3, size: 11pt)[Weights *configurable*\ (default: equal 1/6)]
    ],
    align(center)[
      #text(fill: fe-c, weight: 700, size: 20pt)[1.0]
      #text(fill: txt3, size: 11pt)[Maximum in cohort]
    ],
  )
]

// ─── 6. System Architecture ───
#content-slide([System Architecture])[
  #set text(size: 14pt)
  #grid(
    columns: (55%, 45%),
    column-gutter: 14pt,
    [
      #block(fill: raised, radius: 10pt, inset: 12pt, width: 100%)[
        #text(size: 12pt, fill: txt2)[
          ```
          ┌─────────────────────────────────────┐
          │       ocm (single binary)           │
          │                                     │
          │  ┌──────┐  ┌────────┐  ┌────────┐  │
          │  │ CLI  │─▶│Pipeline│─▶│ SQLite │  │
          │  └──────┘  │        │  └────────┘  │
          │            │ Parser │      │       │
          │            │ GitLog │      ▼       │
          │            │ Metrics│  ┌────────┐  │
          │            │ Score  │  │HTTP API│  │
          │            └────────┘  │+Dashbd │  │
          │                        └────────┘  │
          └─────────────────────────────────────┘
          ```
        ]
      ]
    ],
    [
      #grid(
        columns: 1fr,
        row-gutter: 8pt,
        block(fill: surface, radius: 8pt, inset: 8pt, width: 100%)[
          #text(fill: accent, weight: 700, size: 12pt)[2 Dependencies Only]
          #v(1pt)
          #text(fill: txt2, size: 12pt)[`yaml.v3` + `modernc.org/sqlite`]
        ],
        block(fill: surface, radius: 8pt, inset: 8pt, width: 100%)[
          #text(fill: dd-c, weight: 700, size: 12pt)[No CGO]
          #v(1pt)
          #text(fill: txt2, size: 12pt)[Pure Go SQLite, cross-compilable]
        ],
        block(fill: surface, radius: 8pt, inset: 8pt, width: 100%)[
          #text(fill: cdr-c, weight: 700, size: 12pt)[Embedded Assets]
          #v(1pt)
          #text(fill: txt2, size: 12pt)[`go:embed` HTML/CSS/JS — zero external files]
        ],
        block(fill: surface, radius: 8pt, inset: 8pt, width: 100%)[
          #text(fill: cv-c, weight: 700, size: 12pt)[7 Packages]
          #v(1pt)
          #text(fill: txt2, size: 12pt)[model, parser, pipeline, gitlog, storage, api, dashboard]
        ],
      )
    ]
  )
]

// ─── 7. Data Flow Pipeline ───
#content-slide([Data Flow Pipeline])[
  #set text(size: 15pt)
  #step-list(color: accent, (
    [*Discovery* — Walk filesystem, collect `.yaml`/`.yml`],
    [*Parsing* — Multi-doc YAML decode, extract CSA + deps + FE],
    [*Service ID* — Map files to services via `dir` or `manifest`],
    [*Metric Engine* — Compute DD, DB, CV, FE, CDR],
    [*Normalization* — Min-max across cohort],
    [*Composite* — Weighted sum to single OCM score],
    [*Persist* — Atomic transaction to SQLite],
    [*Serve* — HTTP API + embedded dashboard],
  ))
]

// ─── 8. Parser & Extraction ───
#content-slide([Parser & Extraction])[
  #set text(size: 12pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 12pt,
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: csa-c, weight: 700, size: 12pt)[CSA — Config Surface Contributors]
        #v(2pt)
        #table(
          columns: (auto, 1fr),
          stroke: 0.4pt + border,
          inset: 4pt,
          fill: (x, y) => if y == 0 { raised } else { none },
          [*Component*], [*Source*],
          [`env`], [Container environment variables],
          [`port`], [Container port entries],
          [`resource`], [Resource limits/requests keys],
          [`replica`], [`spec.replicas` field],
          [`spec_key`], [Top-level spec keys],
        )
      ]

      #v(4pt)

      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: fe-c, weight: 700, size: 12pt)[FE — 5 Detection Rules]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - LoadBalancer / NodePort ports
        - Ingress `paths[]` entries
        - ExternalName services
        - Container `hostPort` bindings
        - External URLs in env vars
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: dd-c, weight: 700, size: 12pt)[Dependency Detection — 3 Heuristics]
        #v(4pt)

        #block(fill: surface, radius: 6pt, inset: 6pt, width: 100%)[
          #text(fill: dd-c, weight: 600, size: 11pt)[Rule 1 — Env var suffix matching]
          #v(1pt)
          #text(size: 11pt, fill: txt3)[`_SERVICE`, `_HOST`, `_ADDR`, `_URL`, `_ENDPOINT`, `_URI`]
        ]

        #v(4pt)

        #block(fill: surface, radius: 6pt, inset: 6pt, width: 100%)[
          #text(fill: db-c, weight: 600, size: 11pt)[Rule 2 — Cluster FQDN]
          #v(1pt)
          #text(size: 11pt, fill: txt3)[Values containing `.svc.cluster.local`]
        ]

        #v(4pt)

        #block(fill: surface, radius: 6pt, inset: 6pt, width: 100%)[
          #text(fill: cdr-c, weight: 600, size: 11pt)[Rule 3 — Hostname:port regex]
          #v(1pt)
          #text(size: 11pt, fill: txt3)[`^[a-z][a-z0-9-]{0,62}:\d{2,5}$`]
        ]
      ]

      #v(4pt)

      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 12pt)[CDR — Config Drift Risk]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        Counts env-specific overrides across: dev, staging, prod, qa, test, uat, preview, canary, local
      ]
    ]
  )
]

// ─── 9. Graph Algorithms ───
#content-slide([Graph Algorithms — Dependency Depth])[
  #set text(size: 14pt)

  #grid(
    columns: (1fr, 1fr),
    column-gutter: 14pt,
    [
      #block(fill: raised, radius: 8pt, inset: 10pt, width: 100%)[
        #text(fill: fe-c, weight: 700, size: 13pt)[Problem]
        #v(2pt)
        #text(fill: txt2, size: 13pt)[Dependency graphs may have cycles (A -> B -> A).]
      ]

      #v(6pt)

      #block(fill: raised, radius: 8pt, inset: 10pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 13pt)[Solution — SCC Condensation]
        #v(2pt)
        #set text(size: 13pt, fill: txt2)
        #step-list(color: dd-c, (
          [Tarjan's algorithm finds SCCs],
          [Condense each SCC to a single node],
          [Result is a DAG (no cycles)],
          [DFS with memoization -> longest path],
        ))
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 10pt, width: 100%)[
        #text(fill: dd-c, weight: 700, size: 13pt)[Example]
        #v(2pt)
        #text(fill: txt2, size: 12pt)[
          ```
          A → B → C    (B ↔ A = cycle)
              ↑ ↙
              A

          SCC: {A, B} → node S₁
          DAG: S₁ → C
          DD[A] = DD[B] = 1,  DD[C] = 0
          ```
        ]
      ]

      #v(4pt)
      #text(size: 11pt, fill: txt3)[Deterministic — same input always produces same DD values.]
    ]
  )
]

// ─── 10. Storage & API ───
#content-slide([Storage & REST API])[
  #set text(size: 13pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 12pt,
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 12pt)[SQLite Schema — 4 Tables]
        #v(2pt)
        #text(fill: txt2, size: 11pt)[
          ```sql
          services(id, name, repository)
          metrics(id, service_id, metric_type,
                  metric_value, timestamp)
          composite_scores(id, service_id,
                  ocm_score, timestamp)
          metric_evidence(id, service_id,
                  metric_type, component,
                  evidence_key, evidence_value,
                  source_path, manifest_kind,
                  manifest_name, timestamp)
          ```
        ]
      ]
      #v(4pt)
      #grid(
        columns: (1fr, 1fr),
        column-gutter: 6pt,
        block(fill: surface, radius: 6pt, inset: 5pt, width: 100%)[
          #text(fill: txt3, size: 10pt)[FK constraints + indexes\ Atomic transactions]
        ],
        block(fill: surface, radius: 6pt, inset: 5pt, width: 100%)[
          #text(fill: txt3, size: 10pt)[JSON responses\ CORS middleware]
        ],
      )
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: dd-c, weight: 700, size: 12pt)[REST API — 6 Endpoints]
        #v(2pt)
        #table(
          columns: (1fr, 1fr),
          stroke: 0.4pt + border,
          inset: 5pt,
          fill: (x, y) => if y == 0 { raised } else { none },
          [*Endpoint*], [*Returns*],
          [`/api/healthz`], [Health check],
          [`/api/services`], [All services],
          [`/api/overview`], [Repo aggregates],
          [`/services/{id}/scores`], [Score time series],
          [`/services/{id}/metrics/{t}`], [Metric series],
          [`/services/{id}/.../evidence`], [Evidence items],
        )
      ]
    ]
  )
]

// ─── 11. Dashboard — Design & Overview Screenshot ───
#content-slide([Dashboard — Design & Overview])[
  #set text(size: 13pt)
  #grid(
    columns: (40%, 60%),
    column-gutter: 12pt,
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 12pt)[Views & Charts]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - Composite OCM score hero card
        - 6 metric tiles with color coding
        - OCM trend chart (Canvas)
        - Radar chart + Bar chart
        - Service comparison table
        - Evidence drill-down modal
      ]

      #v(4pt)

      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 12pt)[Tech Stack]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - Stripe-inspired dark theme
        - Pure HTML/CSS/JS (no frameworks)
        - Canvas-based chart rendering
        - Responsive + embedded via `go:embed`
      ]
    ],
    [
      #align(center)[
        #block(radius: 8pt, clip: true, stroke: 0.5pt + border, width: 100%)[
          #image("screenshots/dashboard-overview.png", width: 100%, height: 72%)
        ]
        #text(size: 10pt, fill: txt3, style: "italic")[Repository overview — composite score, 9 services, 6 metric tiles]
      ]
    ]
  )
]

// ─── 12. Dashboard — Service Detail & Charts ───
#content-slide([Dashboard — Service Detail & Evidence])[
  #set text(size: 13pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 12pt,
    [
      #align(center)[
        #block(radius: 8pt, clip: true, stroke: 0.5pt + border, width: 100%)[
          #image("screenshots/dashboard-service.png", width: 100%, height: 290pt)
        ]
        #text(size: 10pt, fill: txt3, style: "italic")[Service detail view]
      ]
    ],
    [
      #align(center)[
        #block(radius: 8pt, clip: true, stroke: 0.5pt + border, width: 100%)[
          #image("screenshots/dashboard-evidence.png", width: 100%, height: 290pt)
        ]
        #text(size: 10pt, fill: txt3, style: "italic")[Evidence drill-down]
      ]
    ]
  )
]

// ─── 13. Evidence Traceability ───
#content-slide([Evidence-Based Traceability])[
  #set text(size: 13pt)

  #table(
    columns: (auto, 1fr, 1fr),
    stroke: 0.5pt + border,
    inset: 7pt,
    fill: (x, y) => if y == 0 { raised } else if calc.odd(y) { surface } else { none },
    [*Metric*], [*Evidence Type*], [*Example*],
    [#metric-pill("CSA", csa-c)], [`env`, `port`, `resource`, `replica`], [`PORT=8080` from `api/deploy.yaml`],
    [#metric-pill("DD", dd-c)], [`dep_chain`, `direct_dep`], [`api -> cache -> db` (depth=2)],
    [#metric-pill("DB", db-c)], [`upstream_dep`, `downstream_dep`], [`orders -> payment` edge],
    [#metric-pill("CV", cv-c)], [`commit`], [`commit abc12345`],
    [#metric-pill("FE", fe-c)], [`loadbalancer`, `ingress`, `external_url`], [LoadBalancer port 80],
    [#metric-pill("CDR", cdr-c)], [`env_override`], [`config.yaml` in dev + prod],
  )

  #v(4pt)
  #callout(color: accent2)[Every metric value links to specific YAML files, manifest keys, and configuration parameters.]
]

// ─── 14. Testing & Validation ───
#content-slide([Testing & Validation])[
  #set text(size: 13pt)
  #grid(
    columns: (1fr, 1fr, 1fr),
    column-gutter: 10pt,
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 12pt)[Parser Tests (16)]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - 6 FE detection tests
        - 8 dependency extraction tests
        - CronJob, FQDN, regex patterns
        - Numeric value rejection
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: dd-c, weight: 700, size: 12pt)[Pipeline Tests (8)]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - DD cycle handling (Tarjan's)
        - DB in/out degree counting
        - Self-dependency ignored
        - Normalization edge cases
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: cdr-c, weight: 700, size: 12pt)[Storage & Git (8)]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - Round-trip all 6 metric types
        - 6 git log parsing tests
        - Empty output handling
        - Graceful degradation
      ]
    ],
  )

  #v(6pt)
  #callout(color: success)[All 32 tests pass #h(12pt) | #h(12pt) Build clean #h(12pt) | #h(12pt) Vet clean]

  #v(4pt)
  #grid(
    columns: (1fr, 1fr, 1fr),
    column-gutter: 8pt,
    block(fill: surface, radius: 6pt, inset: 6pt, width: 100%)[
      #text(fill: txt3, size: 11pt)[*Unit:* Isolated function tests\ (normalization, extraction)]
    ],
    block(fill: surface, radius: 6pt, inset: 6pt, width: 100%)[
      #text(fill: txt3, size: 11pt)[*Integration:* Full pipeline\ with temp K8s repos]
    ],
    block(fill: surface, radius: 6pt, inset: 6pt, width: 100%)[
      #text(fill: txt3, size: 11pt)[*Persistence:* Round-trip\ in-memory SQLite]
    ],
  )
]

// ─── 15. Phase Evolution & Limitations ───
#content-slide([Phase Evolution & Limitations])[
  #set text(size: 13pt)
  #grid(
    columns: (1fr, 1fr, 1fr),
    column-gutter: 10pt,
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: txt3, weight: 700, size: 12pt)[Phase 2 — MVP]
        #v(2pt)
        #set text(size: 12pt, fill: txt3)
        - 2 metrics: CSA + DD
        - Basic dependency extraction
        - Simple dashboard
        - CSA evidence only
        - 50/50 composite weights
      ]
    ],
    [
      #block(fill: accent.lighten(92%), radius: 8pt, inset: 8pt, width: 100%, stroke: 0.5pt + accent.lighten(70%))[
        #text(fill: accent, weight: 700, size: 12pt)[Phase 3 — Complete]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - All 6 metrics implemented
        - 9 dependency patterns
        - Git integration for CV
        - Evidence for all metrics
        - Redesigned dashboard
        - 32 automated tests
      ]
    ],
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: cv-c, weight: 700, size: 12pt)[Known Limitations]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - No Helm template rendering
        - Heuristic-based extraction
        - Single-run model
        - `git` binary needed for CV
        - Localhost only (no auth)
      ]
    ],
  )

  #v(4pt)

  #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
    #text(fill: cdr-c, weight: 700, size: 12pt)[Mitigations]
    #h(10pt)
    #text(size: 12pt, fill: txt2)[
      Evidence trail for verification #h(6pt) | #h(6pt)
      Multiple service ID strategies #h(6pt) | #h(6pt)
      Best-effort approach #h(6pt) | #h(6pt)
      Modular architecture for extension
    ]
  ]
]

// ─── 16. Future Work & Summary ───
#content-slide([Future Work & Summary])[
  #set text(size: 13pt)
  #grid(
    columns: (1fr, 1fr),
    column-gutter: 14pt,
    [
      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: accent, weight: 700, size: 12pt)[Near-Term Extensions]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - Helm v3 template rendering
        - Watch mode / scheduled re-analysis
        - Configurable weights via UI
        - Export reports (PDF/CSV)
      ]

      #v(4pt)

      #block(fill: raised, radius: 8pt, inset: 8pt, width: 100%)[
        #text(fill: accent2, weight: 700, size: 12pt)[Long-Term Vision]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        - CI/CD integration (PR complexity diffs)
        - Multi-repo support
        - Complexity budget enforcement
        - Authentication & multi-user
      ]
    ],
    [
      #block(fill: success.lighten(92%), radius: 8pt, inset: 8pt, width: 100%, stroke: 0.5pt + success.lighten(70%))[
        #text(fill: success, weight: 700, size: 12pt)[Key Achievements]
        #v(2pt)
        #set text(size: 12pt, fill: txt2)
        #grid(
          columns: (auto, 1fr),
          column-gutter: 6pt,
          row-gutter: 5pt,
          text(fill: success, weight: 800)[+], [*6 metrics* quantifying ops complexity],
          text(fill: success, weight: 800)[+], [*Single binary* — Go + SQLite + dashboard],
          text(fill: success, weight: 800)[+], [*Evidence-based* — full traceability],
          text(fill: success, weight: 800)[+], [*Graph algorithms* — Tarjan's SCC],
          text(fill: success, weight: 800)[+], [*2 dependencies* — yaml.v3 + pure-Go SQLite],
          text(fill: success, weight: 800)[+], [*Interactive dashboard* — charts & drill-down],
          text(fill: success, weight: 800)[+], [*32 automated tests* across 4 packages],
        )
      ]
    ]
  )
]

// ─── 17. Thank You ───
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
    #v(12pt)
    #text(size: 12pt, fill: accent)[
      #link("https://github.com/aabidsofi19/ocm")[github.com/aabidsofi19/ocm]
    ]
  ]
]
