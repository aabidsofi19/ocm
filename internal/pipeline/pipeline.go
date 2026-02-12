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

	"ocm/internal/model"
	"ocm/internal/parser"
)

type Options struct {
	RepoPath           string
	ServiceKeyStrategy string // dir|manifest
	Now                func() time.Time
	Logger             *log.Logger
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
	for i := range services {
		if services[i].Metrics == nil {
			services[i].Metrics = map[model.MetricType]float64{}
		}
		services[i].Metrics[model.MetricDD] = float64(dd[services[i].Name])
	}

	// Normalize across cohort for this run.
	normCSA := normalize(services, model.MetricCSA)
	normDD := normalize(services, model.MetricDD)
	for i := range services {
		if services[i].Normalized == nil {
			services[i].Normalized = map[model.MetricType]float64{}
		}
		services[i].Normalized[model.MetricCSA] = normCSA[services[i].Name]
		services[i].Normalized[model.MetricDD] = normDD[services[i].Name]

		// MVP composite: only CSA + DD. Default weights: 0.5/0.5; missing treated as 0.
		services[i].Score = 0.5*services[i].Normalized[model.MetricCSA] + 0.5*services[i].Normalized[model.MetricDD]
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
		deps     map[string]struct{}
		evidence []model.EvidenceItem
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
			acc = &svcAccum{name: serviceName, repo: root, deps: map[string]struct{}{}}
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
			for _, dep := range f.Dependencies {
				acc.deps[dep] = struct{}{}
			}
			for _, ev := range f.Evidence {
				acc.evidence = append(acc.evidence, model.EvidenceItem{
					MetricType:   model.MetricCSA,
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
				for dep := range acc.deps {
					macc.deps[dep] = struct{}{}
				}
				macc.evidence = append(macc.evidence, acc.evidence...)
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
		m := map[model.MetricType]float64{model.MetricCSA: acc.csa}
		out = append(out, model.AnalysisServiceResult{
			Name:         acc.name,
			Repository:   root,
			Dependencies: mapKeys(acc.deps),
			Metrics:      m,
			Evidence:     map[model.MetricType][]model.EvidenceItem{model.MetricCSA: acc.evidence},
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
