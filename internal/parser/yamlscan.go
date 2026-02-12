package parser

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

type K8sDocFacts struct {
	ManifestName string
	ManifestKind string
	CSA          int
	Dependencies []string
	Evidence     []EvidenceItem
}

type EvidenceItem struct {
	Component    string
	Key          string
	Value        string
	SourcePath   string
	ManifestKind string
	ManifestName string
}

func ParseK8sYAMLForFacts(doc []byte, sourcePath string) ([]K8sDocFacts, error) {
	dec := yaml.NewDecoder(bytes.NewReader(doc))

	var out []K8sDocFacts
	for {
		var n any
		err := dec.Decode(&n)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode yaml: %w", err)
		}
		m, ok := n.(map[string]any)
		if !ok {
			continue
		}
		facts := K8sDocFacts{}
		facts.ManifestName = getString(m, "metadata", "name")
		facts.ManifestKind = getString(m, "kind")

		// heuristic configurable parameters: count keys in spec top-level
		if spec, ok := getMap(m, "spec"); ok {
			facts.CSA += len(spec)
			for k := range spec {
				facts.Evidence = append(facts.Evidence, EvidenceItem{
					Component:    "spec_key",
					Key:          k,
					SourcePath:   sourcePath,
					ManifestKind: facts.ManifestKind,
					ManifestName: facts.ManifestName,
				})
			}
		}
		// replicas
		if v, ok := getNumber(m, "spec", "replicas"); ok {
			facts.CSA++
			facts.Evidence = append(facts.Evidence, EvidenceItem{
				Component:    "replica",
				Key:          "spec.replicas",
				Value:        fmt.Sprintf("%v", v),
				SourcePath:   sourcePath,
				ManifestKind: facts.ManifestKind,
				ManifestName: facts.ManifestName,
			})
		}

		containers := getContainers(m)
		for _, c := range containers {
			// env vars
			if env, ok := c["env"].([]any); ok {
				facts.CSA += len(env)
				for _, ev := range env {
					em, _ := ev.(map[string]any)
					k := strings.ToUpper(getString(em, "name"))
					val := getString(em, "value")
					facts.Evidence = append(facts.Evidence, EvidenceItem{
						Component:    "env",
						Key:          k,
						Value:        val,
						SourcePath:   sourcePath,
						ManifestKind: facts.ManifestKind,
						ManifestName: facts.ManifestName,
					})
					if strings.HasSuffix(k, "_SERVICE") || strings.HasSuffix(k, "_SERVICE_HOST") {
						dep := normalizeServiceName(val)
						if dep != "" {
							facts.Dependencies = append(facts.Dependencies, dep)
						}
					}
				}
			}
			// resources
			if r, ok := c["resources"].(map[string]any); ok {
				leaves := leafPaths(r, "resources")
				facts.CSA += len(leaves)
				for _, lp := range leaves {
					facts.Evidence = append(facts.Evidence, EvidenceItem{
						Component:    "resource",
						Key:          lp,
						SourcePath:   sourcePath,
						ManifestKind: facts.ManifestKind,
						ManifestName: facts.ManifestName,
					})
				}
			}
			// ports
			if ports, ok := c["ports"].([]any); ok {
				facts.CSA += len(ports)
				for _, pv := range ports {
					pm, _ := pv.(map[string]any)
					cp := getString(pm, "containerPort")
					if cp == "" {
						if n, ok := pm["containerPort"]; ok {
							cp = fmt.Sprintf("%v", n)
						}
					}
					facts.Evidence = append(facts.Evidence, EvidenceItem{
						Component:    "port",
						Key:          "containerPort",
						Value:        cp,
						SourcePath:   sourcePath,
						ManifestKind: facts.ManifestKind,
						ManifestName: facts.ManifestName,
					})
				}
			}
		}

		facts.Dependencies = dedupe(facts.Dependencies)
		out = append(out, facts)
	}

	return out, nil
}

func getContainers(doc map[string]any) []map[string]any {
	// spec.template.spec.containers
	if spec, ok := getMap(doc, "spec"); ok {
		if tmpl, ok := spec["template"].(map[string]any); ok {
			if ts, ok := tmpl["spec"].(map[string]any); ok {
				if cs, ok := ts["containers"].([]any); ok {
					var out []map[string]any
					for _, c := range cs {
						if cm, ok := c.(map[string]any); ok {
							out = append(out, cm)
						}
					}
					return out
				}
			}
		}
	}
	return nil
}

func leafPaths(m map[string]any, prefix string) []string {
	var out []string
	for k, v := range m {
		p := k
		if prefix != "" {
			p = prefix + "." + k
		}
		if mm, ok := v.(map[string]any); ok {
			out = append(out, leafPaths(mm, p)...)
			continue
		}
		out = append(out, p)
	}
	return out
}

func getString(m map[string]any, keys ...string) string {
	cur := any(m)
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur, ok = mm[k]
		if !ok {
			return ""
		}
	}
	s, _ := cur.(string)
	return s
}

func getMap(m map[string]any, key string) (map[string]any, bool) {
	v, ok := m[key]
	if !ok {
		return nil, false
	}
	mm, ok := v.(map[string]any)
	return mm, ok
}

func getNumber(m map[string]any, keys ...string) (float64, bool) {
	cur := any(m)
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return 0, false
		}
		cur, ok = mm[k]
		if !ok {
			return 0, false
		}
	}
	switch v := cur.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
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
	return out
}

func normalizeServiceName(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "http://")
	v = strings.TrimPrefix(v, "https://")
	if v == "" {
		return ""
	}
	// host:port -> host
	if i := strings.IndexByte(v, ':'); i >= 0 {
		v = v[:i]
	}
	// foo.bar.svc.cluster.local -> foo
	if i := strings.IndexByte(v, '.'); i >= 0 {
		v = v[:i]
	}
	return v
}
