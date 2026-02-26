package parser

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// depSuffixes are env var name suffixes that commonly indicate a reference to
// another service. The list is intentionally broad to catch real-world patterns
// across Sock Shop, Google Online Boutique, Helm charts, and typical
// Rails/Django/Spring deployments.
//
// Note: _SERVICE_PORT is excluded because its value is a port number, not a
// service reference.
var depSuffixes = []string{
	"_SERVICE", "_SERVICE_HOST",
	"_ADDR",
	"_HOST",
	"_URL",
	"_ENDPOINT",
	"_URI",
}

// serviceAddrRe matches values that look like a Kubernetes service DNS name
// followed by a port, e.g. "user-db:27017", "redis-cart:6379".
// Valid K8s service names: lowercase alphanumeric + hyphens, 1-63 chars.
var serviceAddrRe = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}:\d{2,5}$`)

type K8sDocFacts struct {
	ManifestName string
	ManifestKind string
	CSA          int
	Dependencies []string
	Evidence     []EvidenceItem
	// FE (Failure Exposure) facts: exposed endpoints and external integrations.
	ExposedEndpoints int
	FEEvidence       []EvidenceItem
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
					if dep := extractDep(k, val); dep != "" {
						facts.Dependencies = append(facts.Dependencies, dep)
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

		// ── FE (Failure Exposure) extraction ──────────────────────────────
		// Spec: FE = # exposed endpoints + external integrations.
		//
		// Detection rules (documented per spec requirement):
		//   1. Service kind with type LoadBalancer or NodePort → 1 exposed endpoint per port.
		//   2. Ingress kind → 1 exposed endpoint per rule/path.
		//   3. Service kind with type ExternalName → 1 external integration.
		//   4. Container ports with hostPort set → 1 exposed endpoint each.
		//   5. Env vars referencing external URLs (http(s)://) that do NOT point
		//      to internal .svc.cluster.local → 1 external integration each.
		extractFE(m, &facts, sourcePath)

		facts.Dependencies = dedupe(facts.Dependencies)
		out = append(out, facts)
	}

	return out, nil
}

func getContainers(doc map[string]any) []map[string]any {
	spec, ok := getMap(doc, "spec")
	if !ok {
		return nil
	}

	// CronJob: spec.jobTemplate.spec.template.spec.containers
	if jt, ok := spec["jobTemplate"].(map[string]any); ok {
		if jtSpec, ok := jt["spec"].(map[string]any); ok {
			return extractContainersFromPodSpec(jtSpec)
		}
	}

	// Deployment/StatefulSet/DaemonSet/Job: spec.template.spec.containers
	return extractContainersFromPodSpec(spec)
}

// extractContainersFromPodSpec extracts containers from a spec that has
// template.spec.containers (works for Deployment, StatefulSet, DaemonSet, Job).
func extractContainersFromPodSpec(spec map[string]any) []map[string]any {
	tmpl, ok := spec["template"].(map[string]any)
	if !ok {
		return nil
	}
	ts, ok := tmpl["spec"].(map[string]any)
	if !ok {
		return nil
	}
	cs, ok := ts["containers"].([]any)
	if !ok {
		return nil
	}
	var out []map[string]any
	for _, c := range cs {
		if cm, ok := c.(map[string]any); ok {
			out = append(out, cm)
		}
	}
	return out
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

// extractDep returns a dependency service name if the given env var (key+value)
// looks like a reference to another Kubernetes service. Returns "" if no
// dependency is detected.
//
// Detection rules (documented per spec requirement):
//
//  1. Env var name ends with a known dependency suffix (_SERVICE, _SERVICE_HOST,
//     _ADDR, _HOST, _URL, _ENDPOINT, _URI). The value is normalized to extract
//     the service hostname.
//  2. Env var value matches hostname:port pattern (e.g. "redis-cart:6379"),
//     indicating a direct service reference regardless of the env var name.
//  3. Env var value contains ".svc.cluster.local", indicating an explicit
//     in-cluster service FQDN reference.
func extractDep(envKey, envVal string) string {
	// Rule 1: suffix-based detection
	for _, suffix := range depSuffixes {
		if strings.HasSuffix(envKey, suffix) {
			dep := normalizeServiceName(envVal)
			if dep != "" {
				return dep
			}
		}
	}

	// Rule 3: explicit cluster FQDN in value (check before hostname:port to
	// get the more specific match).
	if strings.Contains(envVal, ".svc.cluster.local") {
		dep := normalizeServiceName(envVal)
		if dep != "" {
			return dep
		}
	}

	// Rule 2: value looks like hostname:port
	if serviceAddrRe.MatchString(envVal) {
		dep := normalizeServiceName(envVal)
		if dep != "" {
			return dep
		}
	}

	return ""
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
	// Reject purely numeric results (e.g. port numbers like "5432").
	if isNumeric(v) {
		return ""
	}
	// Reject localhost — it's a self-reference, not a cross-service dependency.
	if v == "localhost" || v == "127" {
		return ""
	}
	return v
}

// isNumeric returns true if s consists entirely of ASCII digits.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// extractFE populates Failure Exposure facts on the given K8sDocFacts.
//
// Detection rules (per spec: FE extraction rules MUST be explicitly documented):
//
//  1. Kind=Service, type=LoadBalancer or NodePort → each port counts as 1 exposed endpoint.
//  2. Kind=Ingress → each rule/path counts as 1 exposed endpoint.
//  3. Kind=Service, type=ExternalName → 1 external integration (the externalName target).
//  4. Container hostPort → 1 exposed endpoint each.
//  5. Container env var whose value starts with http:// or https:// and does NOT
//     contain ".svc.cluster.local" → 1 external integration.
func extractFE(m map[string]any, facts *K8sDocFacts, sourcePath string) {
	kind := strings.ToLower(getString(m, "kind"))

	switch kind {
	case "service":
		svcType := getString(m, "spec", "type")
		switch strings.ToLower(svcType) {
		case "loadbalancer", "nodeport":
			ports := getSlice(m, "spec", "ports")
			if len(ports) == 0 {
				// At least 1 exposed endpoint for the service itself.
				facts.ExposedEndpoints++
				facts.FEEvidence = append(facts.FEEvidence, EvidenceItem{
					Component:    "exposed_service",
					Key:          fmt.Sprintf("Service/%s", facts.ManifestName),
					Value:        svcType,
					SourcePath:   sourcePath,
					ManifestKind: facts.ManifestKind,
					ManifestName: facts.ManifestName,
				})
			}
			for _, p := range ports {
				pm, ok := p.(map[string]any)
				if !ok {
					continue
				}
				portNum := ""
				if n, ok := pm["port"]; ok {
					portNum = fmt.Sprintf("%v", n)
				}
				facts.ExposedEndpoints++
				facts.FEEvidence = append(facts.FEEvidence, EvidenceItem{
					Component:    "exposed_port",
					Key:          fmt.Sprintf("Service/%s:%s", facts.ManifestName, portNum),
					Value:        svcType,
					SourcePath:   sourcePath,
					ManifestKind: facts.ManifestKind,
					ManifestName: facts.ManifestName,
				})
			}
		case "externalname":
			extName := getString(m, "spec", "externalName")
			facts.ExposedEndpoints++
			facts.FEEvidence = append(facts.FEEvidence, EvidenceItem{
				Component:    "external_integration",
				Key:          fmt.Sprintf("ExternalName/%s", facts.ManifestName),
				Value:        extName,
				SourcePath:   sourcePath,
				ManifestKind: facts.ManifestKind,
				ManifestName: facts.ManifestName,
			})
		}

	case "ingress":
		rules := getSlice(m, "spec", "rules")
		if len(rules) == 0 {
			// Ingress with no rules still represents at least 1 exposed endpoint.
			facts.ExposedEndpoints++
			facts.FEEvidence = append(facts.FEEvidence, EvidenceItem{
				Component:    "ingress_endpoint",
				Key:          fmt.Sprintf("Ingress/%s", facts.ManifestName),
				Value:        "(default)",
				SourcePath:   sourcePath,
				ManifestKind: facts.ManifestKind,
				ManifestName: facts.ManifestName,
			})
		}
		for _, rule := range rules {
			rm, ok := rule.(map[string]any)
			if !ok {
				continue
			}
			host := ""
			if h, ok := rm["host"].(string); ok {
				host = h
			}
			httpPaths := getSlice(rm, "http", "paths")
			if len(httpPaths) == 0 {
				facts.ExposedEndpoints++
				facts.FEEvidence = append(facts.FEEvidence, EvidenceItem{
					Component:    "ingress_endpoint",
					Key:          fmt.Sprintf("Ingress/%s", facts.ManifestName),
					Value:        host,
					SourcePath:   sourcePath,
					ManifestKind: facts.ManifestKind,
					ManifestName: facts.ManifestName,
				})
			}
			for _, hp := range httpPaths {
				hpm, ok := hp.(map[string]any)
				if !ok {
					continue
				}
				pathVal := ""
				if pv, ok := hpm["path"].(string); ok {
					pathVal = pv
				}
				facts.ExposedEndpoints++
				facts.FEEvidence = append(facts.FEEvidence, EvidenceItem{
					Component:    "ingress_endpoint",
					Key:          fmt.Sprintf("Ingress/%s%s", host, pathVal),
					Value:        host + pathVal,
					SourcePath:   sourcePath,
					ManifestKind: facts.ManifestKind,
					ManifestName: facts.ManifestName,
				})
			}
		}
	}

	// Container-level FE signals.
	containers := getContainers(m)
	for _, c := range containers {
		// hostPort → exposed endpoint.
		if ports, ok := c["ports"].([]any); ok {
			for _, pv := range ports {
				pm, _ := pv.(map[string]any)
				if hp, ok := pm["hostPort"]; ok {
					facts.ExposedEndpoints++
					facts.FEEvidence = append(facts.FEEvidence, EvidenceItem{
						Component:    "host_port",
						Key:          fmt.Sprintf("hostPort:%v", hp),
						Value:        fmt.Sprintf("%v", hp),
						SourcePath:   sourcePath,
						ManifestKind: facts.ManifestKind,
						ManifestName: facts.ManifestName,
					})
				}
			}
		}
		// Env vars pointing to external URLs.
		if env, ok := c["env"].([]any); ok {
			for _, ev := range env {
				em, _ := ev.(map[string]any)
				val := getString(em, "value")
				if val == "" {
					continue
				}
				if (strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://")) &&
					!strings.Contains(val, ".svc.cluster.local") {
					facts.ExposedEndpoints++
					facts.FEEvidence = append(facts.FEEvidence, EvidenceItem{
						Component:    "external_url",
						Key:          getString(em, "name"),
						Value:        val,
						SourcePath:   sourcePath,
						ManifestKind: facts.ManifestKind,
						ManifestName: facts.ManifestName,
					})
				}
			}
		}
	}
}

// getSlice navigates nested map keys and returns the final value as []any.
func getSlice(m map[string]any, keys ...string) []any {
	cur := any(m)
	for _, k := range keys {
		mm, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = mm[k]
		if !ok {
			return nil
		}
	}
	sl, _ := cur.([]any)
	return sl
}
