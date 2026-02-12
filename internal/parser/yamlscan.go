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
	CSA          int
	Dependencies []string
}

func ParseK8sYAMLForFacts(doc []byte) ([]K8sDocFacts, error) {
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

		// heuristic configurable parameters: count keys in spec top-level
		if spec, ok := getMap(m, "spec"); ok {
			facts.CSA += len(spec)
		}
		// replicas
		if _, ok := getNumber(m, "spec", "replicas"); ok {
			facts.CSA++
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
				facts.CSA += countLeafKeys(r)
			}
			// ports
			if ports, ok := c["ports"].([]any); ok {
				facts.CSA += len(ports)
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

func countLeafKeys(m map[string]any) int {
	count := 0
	for _, v := range m {
		if mm, ok := v.(map[string]any); ok {
			count += countLeafKeys(mm)
			continue
		}
		count++
	}
	return count
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
