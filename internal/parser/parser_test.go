package parser

import (
	"sort"
	"testing"
)

func TestParseFE_ServiceLoadBalancer(t *testing.T) {
	yaml := `
apiVersion: v1
kind: Service
metadata:
  name: my-svc
spec:
  type: LoadBalancer
  ports:
    - port: 80
      targetPort: 8080
    - port: 443
      targetPort: 8443
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(facts))
	}
	f := facts[0]
	if f.ExposedEndpoints != 2 {
		t.Fatalf("expected 2 exposed endpoints (2 LB ports), got %d", f.ExposedEndpoints)
	}
	if len(f.FEEvidence) != 2 {
		t.Fatalf("expected 2 FE evidence items, got %d", len(f.FEEvidence))
	}
}

func TestParseFE_Ingress(t *testing.T) {
	yaml := `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
    - host: example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: api-svc
                port:
                  number: 80
          - path: /web
            pathType: Prefix
            backend:
              service:
                name: web-svc
                port:
                  number: 80
    - host: admin.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: admin-svc
                port:
                  number: 80
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "ingress.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(facts))
	}
	f := facts[0]
	// 2 paths under example.com + 1 path under admin.example.com = 3
	if f.ExposedEndpoints != 3 {
		t.Fatalf("expected 3 exposed endpoints, got %d", f.ExposedEndpoints)
	}
}

func TestParseFE_ExternalName(t *testing.T) {
	yaml := `
apiVersion: v1
kind: Service
metadata:
  name: ext-db
spec:
  type: ExternalName
  externalName: db.external.example.com
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "ext.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f := facts[0]
	if f.ExposedEndpoints != 1 {
		t.Fatalf("expected 1 external integration, got %d", f.ExposedEndpoints)
	}
	if f.FEEvidence[0].Component != "external_integration" {
		t.Fatalf("expected component=external_integration, got %s", f.FEEvidence[0].Component)
	}
}

func TestParseFE_ExternalURL(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: worker
spec:
  template:
    spec:
      containers:
        - name: worker
          env:
            - name: WEBHOOK_URL
              value: "https://hooks.slack.com/services/T00/B00/xxx"
            - name: INTERNAL_SVC
              value: "http://svc.default.svc.cluster.local:8080"
            - name: DB_HOST
              value: "postgres-host"
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "deploy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f := facts[0]
	// Only the Slack URL counts (not the internal svc.cluster.local or plain hostname)
	if f.ExposedEndpoints != 1 {
		t.Fatalf("expected 1 external URL, got %d", f.ExposedEndpoints)
	}
}

func TestParseFE_NodePort(t *testing.T) {
	yaml := `
apiVersion: v1
kind: Service
metadata:
  name: nodeport-svc
spec:
  type: NodePort
  ports:
    - port: 80
      nodePort: 30080
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "np.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f := facts[0]
	if f.ExposedEndpoints != 1 {
		t.Fatalf("expected 1 exposed endpoint for NodePort, got %d", f.ExposedEndpoints)
	}
}

func TestParseFE_ClusterIP_NoExposure(t *testing.T) {
	yaml := `
apiVersion: v1
kind: Service
metadata:
  name: internal-svc
spec:
  type: ClusterIP
  ports:
    - port: 80
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "cip.yaml")
	if err != nil {
		t.Fatal(err)
	}
	f := facts[0]
	if f.ExposedEndpoints != 0 {
		t.Fatalf("ClusterIP should have 0 exposed endpoints, got %d", f.ExposedEndpoints)
	}
}

// ── Dependency extraction tests ──

func TestDep_ServiceSuffix(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
        - name: app
          env:
            - name: DB_SERVICE
              value: "postgres"
            - name: CACHE_SERVICE_HOST
              value: "redis"
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := facts[0].Dependencies
	sort.Strings(deps)
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d: %v", len(deps), deps)
	}
	if deps[0] != "postgres" || deps[1] != "redis" {
		t.Fatalf("expected [postgres redis], got %v", deps)
	}
}

func TestDep_AddrSuffix(t *testing.T) {
	// Google Online Boutique pattern: PRODUCT_CATALOG_SERVICE_ADDR=productcatalogservice:3550
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: checkout
spec:
  template:
    spec:
      containers:
        - name: checkout
          env:
            - name: PRODUCT_CATALOG_SERVICE_ADDR
              value: "productcatalogservice:3550"
            - name: CART_SERVICE_ADDR
              value: "cartservice:7070"
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := facts[0].Dependencies
	sort.Strings(deps)
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d: %v", len(deps), deps)
	}
	if deps[0] != "cartservice" || deps[1] != "productcatalogservice" {
		t.Fatalf("expected [cartservice productcatalogservice], got %v", deps)
	}
}

func TestDep_HostSuffix(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
        - name: app
          env:
            - name: DB_HOST
              value: "postgres:5432"
            - name: REDIS_HOST
              value: "redis"
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := facts[0].Dependencies
	sort.Strings(deps)
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d: %v", len(deps), deps)
	}
	if deps[0] != "postgres" || deps[1] != "redis" {
		t.Fatalf("expected [postgres redis], got %v", deps)
	}
}

func TestDep_URLSuffix(t *testing.T) {
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
        - name: app
          env:
            - name: API_URL
              value: "http://api-gateway:8080/v1"
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := facts[0].Dependencies
	if len(deps) != 1 || deps[0] != "api-gateway" {
		t.Fatalf("expected [api-gateway], got %v", deps)
	}
}

func TestDep_HostnamePortValue(t *testing.T) {
	// Sock Shop pattern: bare env var name with hostname:port value
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user
spec:
  template:
    spec:
      containers:
        - name: user
          env:
            - name: mongo
              value: "user-db:27017"
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := facts[0].Dependencies
	if len(deps) != 1 || deps[0] != "user-db" {
		t.Fatalf("expected [user-db], got %v", deps)
	}
}

func TestDep_ClusterLocalFQDN(t *testing.T) {
	// Sock Shop pattern: ZIPKIN=zipkin.jaeger.svc.cluster.local
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: shipping
spec:
  template:
    spec:
      containers:
        - name: shipping
          env:
            - name: ZIPKIN
              value: "zipkin.jaeger.svc.cluster.local"
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := facts[0].Dependencies
	if len(deps) != 1 || deps[0] != "zipkin" {
		t.Fatalf("expected [zipkin], got %v", deps)
	}
}

func TestDep_NoDeps_PlainString(t *testing.T) {
	// Plain string values that are NOT service references should not produce deps
	yaml := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
        - name: app
          env:
            - name: LOG_LEVEL
              value: "debug"
            - name: JAVA_OPTS
              value: "-Xmx512m"
            - name: SESSION_REDIS
              value: "true"
`
	facts, err := ParseK8sYAMLForFacts([]byte(yaml), "test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if len(facts[0].Dependencies) != 0 {
		t.Fatalf("expected 0 deps for plain values, got %v", facts[0].Dependencies)
	}
}
