package pipeline

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ocm/internal/gitlog"
	"ocm/internal/model"
	"ocm/internal/parser"
)

type Options struct {
	RepoPath           string
	ServiceKeyStrategy string // dir|manifest
	Now                func() time.Time
	Logger             *log.Logger
	// CVWindow is the lookback window for Change Volatility.
	// Default: 30 days.
	CVWindow time.Duration
}

type Pipeline struct {
	opts Options
}

func New(opts Options) *Pipeline {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Logger == nil {
		opts.Logger = log.New(os.Stderr, "", 0)
	}
	if opts.ServiceKeyStrategy == "" {
		opts.ServiceKeyStrategy = "dir"
	}
	return &Pipeline{opts: opts}
}

func (p *Pipeline) Run(ctx context.Context) (model.AnalysisResult, error) {
	_ = ctx

	res := model.AnalysisResult{RunAt: p.opts.Now().UTC(), RepoPath: p.opts.RepoPath}

	services, warnings, err := p.discoverAndParse()
	if err != nil {
		return model.AnalysisResult{}, err
	}
	res.Warnings = warnings

	// DD computed from extracted dependency graph; CSA from config facts.
	dd := computeDD(services)
	db, dbEvidence := computeDB(services)
	cdr, cdrEvidence := computeCDR(p.opts.RepoPath, services)

	// CV: Change Volatility via Git history.
	cvWindow := p.opts.CVWindow
	if cvWindow == 0 {
		cvWindow = gitlog.DefaultWindow
	}
	changeFacts, err := gitlog.ExtractChangeFacts(gitlog.Options{
		RepoPath: p.opts.RepoPath,
		Window:   cvWindow,
		Now:      p.opts.Now,
	})
	if err != nil {
		res.Warnings = append(res.Warnings, fmt.Sprintf("git history: %v", err))
	}

	for i := range services {
		if services[i].Metrics == nil {
			services[i].Metrics = map[model.MetricType]float64{}
		}
		services[i].Metrics[model.MetricDD] = float64(dd[services[i].Name])
		services[i].Metrics[model.MetricDB] = float64(db[services[i].Name])
		services[i].Metrics[model.MetricCDR] = cdr[services[i].Name]

		// CV metric value: commit count within the time window.
		if cf, ok := changeFacts[services[i].Name]; ok {
			services[i].Metrics[model.MetricCV] = float64(cf.CommitCount)
			// CV evidence: list the commits.
			var cvEvidence []model.EvidenceItem
			for _, hash := range cf.Commits {
				cvEvidence = append(cvEvidence, model.EvidenceItem{
					MetricType: model.MetricCV,
					Component:  "commit",
					Key:        hash,
					Value:      fmt.Sprintf("commit %s", hash[:minInt(8, len(hash))]),
				})
			}
			if services[i].Evidence == nil {
				services[i].Evidence = map[model.MetricType][]model.EvidenceItem{}
			}
			if len(cvEvidence) > 0 {
				services[i].Evidence[model.MetricCV] = cvEvidence
			}
		} else {
			services[i].Metrics[model.MetricCV] = 0
		}

		if services[i].Evidence == nil {
			services[i].Evidence = map[model.MetricType][]model.EvidenceItem{}
		}
		if ev, ok := dbEvidence[services[i].Name]; ok && len(ev) > 0 {
			services[i].Evidence[model.MetricDB] = ev
		}
		if ev, ok := cdrEvidence[services[i].Name]; ok && len(ev) > 0 {
			services[i].Evidence[model.MetricCDR] = ev
		}
	}

	// Normalize across cohort for this run.
	normalized := map[model.MetricType]map[string]float64{}
	for _, mt := range model.AllMetrics {
		normalized[mt] = normalize(services, mt)
	}
	for i := range services {
		if services[i].Normalized == nil {
			services[i].Normalized = map[model.MetricType]float64{}
		}
		for _, mt := range model.AllMetrics {
			services[i].Normalized[mt] = normalized[mt][services[i].Name]
		}

		// Composite OCM score: weighted sum of all available normalized metrics.
		// Missing metrics are treated as 0. Weights from model.DefaultWeights.
		score := 0.0
		for _, mt := range model.AllMetrics {
			score += model.DefaultWeights[mt] * services[i].Normalized[mt]
		}
		services[i].Score = score
	}

	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	res.Services = services
	return res, nil
}

// discoverAndParse implements a best-effort Kubernetes YAML scan.
// MVP behavior: treat each directory containing manifests as a service (dir strategy).
// Dependency extraction is heuristic: env vars ending with _SERVICE / _SERVICE_HOST are treated as a dependency.
func (p *Pipeline) discoverAndParse() ([]model.AnalysisServiceResult, []string, error) {
	root := p.opts.RepoPath
	info, err := os.Stat(root)
	if err != nil {
		return nil, nil, err
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("repo path is not a directory: %s", root)
	}

	var warnings []string

	type svcAccum struct {
		name     string
		repo     string
		csa      float64
		fe       float64 // Failure Exposure: exposed endpoints + external integrations
		deps     map[string]struct{}
		evidence map[model.MetricType][]model.EvidenceItem
		srcs     int
		parse    int
	}

	byService := map[string]*svcAccum{}
	maybeAdd := func(serviceName string) *svcAccum {
		if serviceName == "" {
			serviceName = "unknown"
		}
		acc := byService[serviceName]
		if acc == nil {
			acc = &svcAccum{
				name:     serviceName,
				repo:     root,
				deps:     map[string]struct{}{},
				evidence: map[model.MetricType][]model.EvidenceItem{},
			}
			byService[serviceName] = acc
		}
		return acc
	}

	isK8s := func(path string) bool {
		base := strings.ToLower(filepath.Base(path))
		if strings.HasSuffix(base, ".yaml") || strings.HasSuffix(base, ".yml") {
			return true
		}
		return false
	}

	serviceNameFromPath := func(path string) string {
		if p.opts.ServiceKeyStrategy == "manifest" {
			return "" // handled later when we can read metadata.name
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return filepath.Base(filepath.Dir(path))
		}
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) == 0 {
			return filepath.Base(filepath.Dir(path))
		}
		return parts[0]
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("walk error: %s: %v", path, err))
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == ".idea" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isK8s(path) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("read error: %s: %v", path, err))
			return nil
		}
		serviceName := serviceNameFromPath(path)
		acc := maybeAdd(serviceName)
		acc.srcs++
		facts, perr := parser.ParseK8sYAMLForFacts(b, path)
		if perr != nil {
			warnings = append(warnings, fmt.Sprintf("parse warning: %s: %v", path, perr))
			return nil
		}
		acc.parse++
		var manSvcName string
		for _, f := range facts {
			acc.csa += float64(f.CSA)
			acc.fe += float64(f.ExposedEndpoints)
			for _, dep := range f.Dependencies {
				acc.deps[dep] = struct{}{}
			}
			for _, ev := range f.Evidence {
				acc.evidence[model.MetricCSA] = append(acc.evidence[model.MetricCSA], model.EvidenceItem{
					MetricType:   model.MetricCSA,
					Component:    ev.Component,
					Key:          ev.Key,
					Value:        ev.Value,
					SourcePath:   ev.SourcePath,
					ManifestKind: ev.ManifestKind,
					ManifestName: ev.ManifestName,
				})
			}
			for _, ev := range f.FEEvidence {
				acc.evidence[model.MetricFE] = append(acc.evidence[model.MetricFE], model.EvidenceItem{
					MetricType:   model.MetricFE,
					Component:    ev.Component,
					Key:          ev.Key,
					Value:        ev.Value,
					SourcePath:   ev.SourcePath,
					ManifestKind: ev.ManifestKind,
					ManifestName: ev.ManifestName,
				})
			}
			if manSvcName == "" && f.ManifestName != "" {
				manSvcName = f.ManifestName
			}
		}
		if p.opts.ServiceKeyStrategy == "manifest" && manSvcName != "" {
			// Re-key into manifest-derived name.
			if serviceName == "" || serviceName == "unknown" {
				// nothing
			}
			if manSvcName != acc.name {
				// merge into manifest key
				macc := maybeAdd(manSvcName)
				macc.csa += acc.csa
				macc.fe += acc.fe
				for dep := range acc.deps {
					macc.deps[dep] = struct{}{}
				}
				for mt, evs := range acc.evidence {
					macc.evidence[mt] = append(macc.evidence[mt], evs...)
				}
				macc.srcs += acc.srcs
				macc.parse += acc.parse
				delete(byService, acc.name)
			}
		}
		return nil
	})
	if err != nil {
		return nil, warnings, err
	}

	var out []model.AnalysisServiceResult
	for _, acc := range byService {
		m := map[model.MetricType]float64{
			model.MetricCSA: acc.csa,
			model.MetricFE:  acc.fe,
		}
		out = append(out, model.AnalysisServiceResult{
			Name:         acc.name,
			Repository:   root,
			Dependencies: mapKeys(acc.deps),
			Metrics:      m,
			Evidence:     acc.evidence,
		})
	}
	if len(out) == 0 {
		warnings = append(warnings, "no Kubernetes YAML manifests found; MVP parser currently scans *.yml/*.yaml")
	}
	return out, warnings, nil
}

func normalize(services []model.AnalysisServiceResult, metric model.MetricType) map[string]float64 {
	min := 0.0
	max := 0.0
	first := true
	for _, s := range services {
		v, ok := s.Metrics[metric]
		if !ok {
			continue
		}
		if first {
			min, max = v, v
			first = false
			continue
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	out := map[string]float64{}
	for _, s := range services {
		v, ok := s.Metrics[metric]
		if !ok {
			out[s.Name] = 0
			continue
		}
		if max == min {
			out[s.Name] = 0
			continue
		}
		out[s.Name] = (v - min) / (max - min)
	}
	return out
}

// computeCDR computes the Configuration Drift Risk metric for each service.
//
// Detection rules (documented per spec requirement):
//
// CDR counts environment-specific configuration overrides. It works by detecting
// environment-flavored directory patterns in the repo and counting YAML files
// that appear under multiple environment directories for the same service. Each
// additional environment copy of a config file (beyond the first/base) counts as
// one override.
//
// Recognized environment directory name patterns:
//
//	dev, development, staging, stg, production, prod, qa, test, uat,
//	preview, canary, local
//
// These may appear as:
//   - Top-level directories: overlays/dev/, environments/prod/
//   - Kustomize-style: base/ vs overlays/{dev,prod}/
//   - Suffixed directories: myservice-dev/, myservice-prod/
//
// If a service has YAML files in N environments, each unique file basename that
// appears in more than one environment contributes (occurrences - 1) overrides.
func computeCDR(repoPath string, services []model.AnalysisServiceResult) (map[string]float64, map[string][]model.EvidenceItem) {
	cdr := map[string]float64{}
	evidence := map[string][]model.EvidenceItem{}

	for _, s := range services {
		cdr[s.Name] = 0
	}

	envPatterns := map[string]bool{
		"dev": true, "development": true,
		"staging": true, "stg": true,
		"production": true, "prod": true,
		"qa": true, "test": true, "uat": true,
		"preview": true, "canary": true,
		"local": true,
	}

	// For each service, walk its files and group YAML by (envTag, basename).
	// A file path like overlays/dev/deployment.yaml → env="dev", base="deployment.yaml"
	// We look for env indicators in any path component.
	type fileEnv struct {
		env      string
		basename string
		path     string
	}

	for _, svc := range services {
		var envFiles []fileEnv

		_ = filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				name := ""
				if d != nil {
					name = d.Name()
				}
				if name == ".git" || name == "node_modules" || name == "vendor" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}
			rel, err := filepath.Rel(repoPath, path)
			if err != nil {
				return nil
			}
			// Check if this file belongs to this service's directory tree.
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) == 0 {
				return nil
			}
			// Service association: first directory component must match service name
			// (same heuristic as the "dir" strategy).
			if parts[0] != svc.Name {
				return nil
			}

			// Find environment indicator in path components.
			detectedEnv := ""
			for _, part := range parts {
				lower := strings.ToLower(part)
				if envPatterns[lower] {
					detectedEnv = lower
					break
				}
				// Check suffix pattern: "myservice-dev" → "dev"
				for env := range envPatterns {
					if strings.HasSuffix(lower, "-"+env) || strings.HasSuffix(lower, "_"+env) {
						detectedEnv = env
						break
					}
				}
				if detectedEnv != "" {
					break
				}
			}

			if detectedEnv != "" {
				envFiles = append(envFiles, fileEnv{
					env:      detectedEnv,
					basename: filepath.Base(path),
					path:     rel,
				})
			}
			return nil
		})

		if len(envFiles) == 0 {
			continue
		}

		// Group by basename → set of environments.
		byBasename := map[string]map[string]string{} // basename → env → path
		for _, ef := range envFiles {
			if byBasename[ef.basename] == nil {
				byBasename[ef.basename] = map[string]string{}
			}
			byBasename[ef.basename][ef.env] = ef.path
		}

		overrides := 0.0
		for basename, envMap := range byBasename {
			if len(envMap) <= 1 {
				continue
			}
			// Each environment beyond the first is an override.
			overrides += float64(len(envMap) - 1)
			envList := make([]string, 0, len(envMap))
			for env := range envMap {
				envList = append(envList, env)
			}
			sort.Strings(envList)
			evidence[svc.Name] = append(evidence[svc.Name], model.EvidenceItem{
				MetricType: model.MetricCDR,
				Component:  "env_override",
				Key:        basename,
				Value:      fmt.Sprintf("found in %d envs: %s", len(envMap), strings.Join(envList, ", ")),
			})
		}
		cdr[svc.Name] = overrides
	}

	return cdr, evidence
}

// computeDB computes the Dependency Breadth metric for each service.
// DB = # direct upstream (in-degree) + # direct downstream (out-degree) dependencies.
// Upstream means "services that depend on this service" (in-edges).
// Downstream means "services this service depends on" (out-edges).
func computeDB(services []model.AnalysisServiceResult) (map[string]int, map[string][]model.EvidenceItem) {
	outDegree := map[string]int{}
	inDegree := map[string]int{}
	evidence := map[string][]model.EvidenceItem{}

	// Initialize all known services.
	for _, s := range services {
		outDegree[s.Name] = 0
		inDegree[s.Name] = 0
	}

	// Count edges. out-degree: dependencies declared by this service.
	// in-degree: how many other services depend on this one.
	seen := map[string]map[string]bool{} // dedup edges
	for _, s := range services {
		seen[s.Name] = map[string]bool{}
		for _, dep := range s.Dependencies {
			if dep == "" || dep == s.Name {
				continue
			}
			if seen[s.Name][dep] {
				continue
			}
			seen[s.Name][dep] = true
			outDegree[s.Name]++
			inDegree[dep]++
			evidence[s.Name] = append(evidence[s.Name], model.EvidenceItem{
				MetricType: model.MetricDB,
				Component:  "downstream_dep",
				Key:        dep,
				Value:      fmt.Sprintf("%s -> %s", s.Name, dep),
			})
		}
	}

	// Record upstream evidence.
	for _, s := range services {
		for dep := range seen {
			if seen[dep][s.Name] {
				evidence[s.Name] = append(evidence[s.Name], model.EvidenceItem{
					MetricType: model.MetricDB,
					Component:  "upstream_dep",
					Key:        dep,
					Value:      fmt.Sprintf("%s -> %s", dep, s.Name),
				})
			}
		}
	}

	db := map[string]int{}
	for name := range outDegree {
		db[name] = outDegree[name] + inDegree[name]
	}
	return db, evidence
}

func computeDD(services []model.AnalysisServiceResult) map[string]int {
	adj := map[string][]string{}
	for _, s := range services {
		adj[s.Name] = nil
	}
	for _, s := range services {
		for _, dep := range s.Dependencies {
			if dep == "" || dep == s.Name {
				continue
			}
			adj[s.Name] = append(adj[s.Name], dep)
			if _, ok := adj[dep]; !ok {
				adj[dep] = nil
			}
		}
	}
	for k := range adj {
		adj[k] = dedupe(adj[k])
	}

	// SCC-condense into DAG and compute longest path (edges) from each SCC.
	compID, comps := tarjanSCC(adj)
	compAdj := map[int][]int{}
	for u, vs := range adj {
		cu := compID[u]
		for _, v := range vs {
			cv := compID[v]
			if cu == cv {
				continue
			}
			compAdj[cu] = append(compAdj[cu], cv)
		}
	}
	for c := range comps {
		compAdj[c] = dedupeInt(compAdj[c])
	}

	// DP on DAG.
	cache := map[int]int{}
	var dfs func(int) int
	dfs = func(c int) int {
		if v, ok := cache[c]; ok {
			return v
		}
		best := 0
		for _, n := range compAdj[c] {
			cand := 1 + dfs(n)
			if cand > best {
				best = cand
			}
		}
		cache[c] = best
		return best
	}

	dd := map[string]int{}
	for node := range adj {
		dd[node] = dfs(compID[node])
	}
	return dd
}

func mapKeys(m map[string]struct{}) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func dedupe(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func dedupeInt(in []int) []int {
	seen := map[int]struct{}{}
	var out []int
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func tarjanSCC(adj map[string][]string) (map[string]int, [][]string) {
	index := 0
	stack := []string{}
	onStack := map[string]bool{}
	idx := map[string]int{}
	low := map[string]int{}
	compID := map[string]int{}
	var comps [][]string

	for n := range adj {
		idx[n] = -1
	}

	var strongconnect func(v string)
	strongconnect = func(v string) {
		idx[v] = index
		low[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range adj[v] {
			if _, ok := idx[w]; !ok {
				idx[w] = -1
			}
			if idx[w] == -1 {
				strongconnect(w)
				if low[w] < low[v] {
					low[v] = low[w]
				}
			} else if onStack[w] {
				if idx[w] < low[v] {
					low[v] = idx[w]
				}
			}
		}

		if low[v] == idx[v] {
			var comp []string
			for {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[n] = false
				compID[n] = len(comps)
				comp = append(comp, n)
				if n == v {
					break
				}
			}
			comps = append(comps, comp)
		}
	}

	for v := range adj {
		if idx[v] == -1 {
			strongconnect(v)
		}
	}

	return compID, comps
}
