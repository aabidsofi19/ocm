0) Pre-demo setup (do this before recording)
- Pick a repo with Kubernetes YAML split by service folders (recommended):
  - argoproj/argocd-example-apps (then point REPO to its apps/ folder)
- Make sure you have go + just.
- Start from a clean DB:
  - delete your demo DB file so the “latest” numbers are easy to explain.
---
1) Intro (20–30s)
Say:
“Today I’m demoing the Operational Complexity Meter POC. It scans Kubernetes YAML in a repo, computes a configuration complexity metric (CSA), stores results in SQLite, and serves a small dashboard + API. The key thing in this POC is drill-down evidence: you can click a metric and see exactly what contributed to it.”
---
2) Start the app (30–45s)
On terminal (share terminal):
rm -f demo.sqlite
just run REPO="/path/to/argocd-example-apps/apps" DB="demo.sqlite" PORT="8080" SERVICE_KEY="dir"
Say:
“The tool analyzes the repo once on startup, persists the results into SQLite, then starts a local web server.”
---
3) Open the dashboard (30–45s)
In browser (still screensharing):
- Open: http://127.0.0.1:8080/
Point out (quickly):
- “Repo CSA (latest sum)” = aggregate CSA across all services.
- “Repo OCM (latest avg)” = average composite score across services.
- “CSA (latest)” and “OCM (latest)” are for the currently selected service.
---
4) Explain CSA + show per-service values (60–90s)
In the Service dropdown:
- Select 2–3 different services.
Say:
“CSA is Configuration Surface Area. Higher CSA means more operational knobs: env vars, ports, resources, replicas, and other spec-level configuration. This POC computes CSA from Kubernetes YAML and shows it per service, plus an aggregate at repo level.”
---
5) Evidence drill-down (60–90s)
Click the CSA (latest) tile.
Say:
“This is the key POC feature: evidence. Instead of just a number, I can click and see the individual items that contributed—env vars, ports, resource keys, replica config, spec keys—plus the source file path and manifest metadata.”
Scroll the evidence table and call out:
- component (env/port/resource/replica/spec_key)
- manifestKind + manifestName
- sourcePath
---
6) Show the API quickly (45–60s)
Open a couple endpoints in a browser tab (or terminal with curl):
- Repo aggregate:
  - http://127.0.0.1:8080/api/overview
- Services list:
  - http://127.0.0.1:8080/api/services
- Evidence for CSA (replace {id} with a real service id from /api/services):
  - http://127.0.0.1:8080/api/services/{id}/metrics/CSA/evidence
Say:
“The UI is backed by a simple embedded API, which makes it easy to integrate or extend later.”
---
7) Optional “change → recompute” moment (90s)
Do this only if you want a strong “it’s real” moment.
- Open one manifest from the selected service (use the sourcePath shown in evidence).
- Add one env var entry (valid YAML), save.
- Stop and rerun:
# Ctrl+C in the terminal running the server, then:
just run REPO="/path/to/argocd-example-apps/apps" DB="demo.sqlite" PORT="8080" SERVICE_KEY="dir"
Refresh the dashboard, re-select the same service, click CSA again.
Say:
“Now you can see CSA changed, and the evidence list includes the new parameter. This demonstrates the full loop: parse → compute → persist → explain.”
Note: the evidence endpoint shows the latest run’s evidence; the DB can contain multiple runs.
---
8) Close (15–25s)
Say:
“This is a POC focusing on CSA + transparency via evidence. Next steps would be to improve the parsers (Helm, Git), refine metric definitions, and add richer scoring/visuals once the inputs are solid.”
---
