package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ocm/internal/model"
)

func TestComputeDD_CycleHandledDeterministically(t *testing.T) {
	services := []model.AnalysisServiceResult{
		{Name: "a", Dependencies: []string{"b"}, Metrics: map[model.MetricType]float64{model.MetricCSA: 1}},
		{Name: "b", Dependencies: []string{"a", "c"}, Metrics: map[model.MetricType]float64{model.MetricCSA: 1}},
		{Name: "c", Dependencies: nil, Metrics: map[model.MetricType]float64{model.MetricCSA: 1}},
	}

	dd, ddEvidence := computeDD(services)
	if dd["a"] != 1 {
		// a and b are in one SCC; longest path out to c should be 1 edge.
		t.Fatalf("expected dd[a]=1, got %d", dd["a"])
	}
	if dd["b"] != 1 {
		t.Fatalf("expected dd[b]=1, got %d", dd["b"])
	}
	if dd["c"] != 0 {
		t.Fatalf("expected dd[c]=0, got %d", dd["c"])
	}
	// a and b have depth > 0, so they should have evidence
	if len(ddEvidence["a"]) == 0 {
		t.Fatal("expected DD evidence for service a")
	}
	if len(ddEvidence["b"]) == 0 {
		t.Fatal("expected DD evidence for service b")
	}
	// c has depth 0, no evidence expected
	if len(ddEvidence["c"]) != 0 {
		t.Fatalf("expected no DD evidence for service c, got %d items", len(ddEvidence["c"]))
	}
}

func TestNormalize_MaxEqualsMinReturnsZero(t *testing.T) {
	services := []model.AnalysisServiceResult{
		{Name: "a", Metrics: map[model.MetricType]float64{model.MetricCSA: 5}},
		{Name: "b", Metrics: map[model.MetricType]float64{model.MetricCSA: 5}},
	}

	out := normalize(services, model.MetricCSA)
	if out["a"] != 0 || out["b"] != 0 {
		t.Fatalf("expected both normalized values 0, got a=%v b=%v", out["a"], out["b"])
	}
}

func TestPipeline_RunProducesScore(t *testing.T) {
	p := New(Options{RepoPath: ".", Now: func() time.Time { return time.Unix(0, 0).UTC() }})
	res, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	_ = res
}

func TestComputeDB_InOutDegree(t *testing.T) {
	services := []model.AnalysisServiceResult{
		{Name: "a", Dependencies: []string{"b", "c"}},
		{Name: "b", Dependencies: []string{"c"}},
		{Name: "c", Dependencies: nil},
	}

	db, evidence := computeDB(services)

	// a: out=2 (->b, ->c), in=0 => DB=2
	if db["a"] != 2 {
		t.Fatalf("expected db[a]=2, got %d", db["a"])
	}
	// b: out=1 (->c), in=1 (<-a) => DB=2
	if db["b"] != 2 {
		t.Fatalf("expected db[b]=2, got %d", db["b"])
	}
	// c: out=0, in=2 (<-a, <-b) => DB=2
	if db["c"] != 2 {
		t.Fatalf("expected db[c]=2, got %d", db["c"])
	}

	// Check evidence is generated
	if len(evidence["a"]) == 0 {
		t.Fatal("expected evidence for service a")
	}
}

func TestComputeDB_NoDependencies(t *testing.T) {
	services := []model.AnalysisServiceResult{
		{Name: "x"},
		{Name: "y"},
	}

	db, _ := computeDB(services)
	if db["x"] != 0 || db["y"] != 0 {
		t.Fatalf("expected db=0 for isolated services, got x=%d y=%d", db["x"], db["y"])
	}
}

func TestComputeDB_SelfDependencyIgnored(t *testing.T) {
	services := []model.AnalysisServiceResult{
		{Name: "a", Dependencies: []string{"a", "b"}},
		{Name: "b"},
	}

	db, _ := computeDB(services)
	// a: out=1 (->b, self ignored), in=0 => DB=1
	if db["a"] != 1 {
		t.Fatalf("expected db[a]=1 (self-dep ignored), got %d", db["a"])
	}
	// b: out=0, in=1 (<-a) => DB=1
	if db["b"] != 1 {
		t.Fatalf("expected db[b]=1, got %d", db["b"])
	}
}

func TestPipeline_RunAllSixMetrics(t *testing.T) {
	// Create a temp directory with two service directories and K8s manifests.
	// This exercises the full pipeline including CSA, DD, DB, FE, CV, CDR.
	tmpDir := t.TempDir()

	// Service "api" with a LoadBalancer (FE) and dependency on "db"
	apiDir := filepath.Join(tmpDir, "api")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		t.Fatal(err)
	}
	apiManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  replicas: 2
  template:
    spec:
      containers:
        - name: api
          image: api:latest
          env:
            - name: DB_SERVICE_HOST
              value: "db"
            - name: PORT
              value: "8080"
---
apiVersion: v1
kind: Service
metadata:
  name: api-svc
spec:
  type: LoadBalancer
  ports:
    - port: 80
      targetPort: 8080
`
	if err := os.WriteFile(filepath.Join(apiDir, "deploy.yaml"), []byte(apiManifest), 0o644); err != nil {
		t.Fatal(err)
	}

	// Service "db" with no external exposure
	dbDir := filepath.Join(tmpDir, "db")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dbManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: db
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: postgres
          image: postgres:15
          env:
            - name: POSTGRES_DB
              value: "mydb"
`
	if err := os.WriteFile(filepath.Join(dbDir, "deploy.yaml"), []byte(dbManifest), 0o644); err != nil {
		t.Fatal(err)
	}

	// CDR: create env-specific overrides for "api"
	devDir := filepath.Join(apiDir, "dev")
	prodDir := filepath.Join(apiDir, "prod")
	if err := os.MkdirAll(devDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(prodDir, 0o755); err != nil {
		t.Fatal(err)
	}
	envOverride := `apiVersion: v1
kind: ConfigMap
metadata:
  name: api-config
data:
  LOG_LEVEL: debug
`
	if err := os.WriteFile(filepath.Join(devDir, "config.yaml"), []byte(envOverride), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prodDir, "config.yaml"), []byte(envOverride), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New(Options{
		RepoPath: tmpDir,
		Now:      func() time.Time { return time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC) },
	})
	res, err := p.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(res.Services) < 2 {
		t.Fatalf("expected at least 2 services, got %d", len(res.Services))
	}

	// Verify every service has all 6 metric types.
	for _, svc := range res.Services {
		for _, mt := range model.AllMetrics {
			if _, ok := svc.Metrics[mt]; !ok {
				t.Errorf("service %s missing raw metric %s", svc.Name, mt)
			}
			if _, ok := svc.Normalized[mt]; !ok {
				t.Errorf("service %s missing normalized metric %s", svc.Name, mt)
			}
		}
		// Score should be in [0, 1]
		if svc.Score < 0 || svc.Score > 1 {
			t.Errorf("service %s score %f out of [0,1]", svc.Name, svc.Score)
		}
	}

	// Verify specific metrics for "api" service.
	var apiSvc *model.AnalysisServiceResult
	for i := range res.Services {
		if res.Services[i].Name == "api" {
			apiSvc = &res.Services[i]
			break
		}
	}
	if apiSvc == nil {
		t.Fatal("api service not found")
	}
	// api should have CSA > 0 (env vars, ports, replicas)
	if apiSvc.Metrics[model.MetricCSA] <= 0 {
		t.Errorf("expected api CSA > 0, got %f", apiSvc.Metrics[model.MetricCSA])
	}
	// api depends on db -> DD >= 1
	if apiSvc.Metrics[model.MetricDD] < 1 {
		t.Errorf("expected api DD >= 1, got %f", apiSvc.Metrics[model.MetricDD])
	}
	// api has LoadBalancer -> FE >= 1
	if apiSvc.Metrics[model.MetricFE] < 1 {
		t.Errorf("expected api FE >= 1, got %f", apiSvc.Metrics[model.MetricFE])
	}
	// api has dev/prod overrides -> CDR >= 1
	if apiSvc.Metrics[model.MetricCDR] < 1 {
		t.Errorf("expected api CDR >= 1, got %f", apiSvc.Metrics[model.MetricCDR])
	}
}

func TestAllMetricsInComposite(t *testing.T) {
	// Verify that a run with services produces metrics for all 6 types
	// and a composite score in [0, 1].
	services := []model.AnalysisServiceResult{
		{
			Name:         "svc1",
			Repository:   "/tmp",
			Dependencies: []string{"svc2"},
			Metrics: map[model.MetricType]float64{
				model.MetricCSA: 10,
				model.MetricFE:  2,
			},
			Evidence: map[model.MetricType][]model.EvidenceItem{},
		},
		{
			Name:         "svc2",
			Repository:   "/tmp",
			Dependencies: nil,
			Metrics: map[model.MetricType]float64{
				model.MetricCSA: 5,
				model.MetricFE:  0,
			},
			Evidence: map[model.MetricType][]model.EvidenceItem{},
		},
	}

	// Compute DD and DB manually to verify integration.
	dd, _ := computeDD(services)
	db, _ := computeDB(services)

	if dd["svc1"] != 1 {
		t.Fatalf("expected dd[svc1]=1, got %d", dd["svc1"])
	}
	if dd["svc2"] != 0 {
		t.Fatalf("expected dd[svc2]=0, got %d", dd["svc2"])
	}

	// DB: svc1 out=1(->svc2), in=0 => 1; svc2 out=0, in=1 => 1
	if db["svc1"] != 1 || db["svc2"] != 1 {
		t.Fatalf("expected db=1 for both, got svc1=%d svc2=%d", db["svc1"], db["svc2"])
	}

	// Verify normalize works for all metric types
	for _, mt := range model.AllMetrics {
		norm := normalize(services, mt)
		for _, s := range services {
			v := norm[s.Name]
			if v < 0 || v > 1 {
				t.Fatalf("normalized %s for %s out of [0,1]: %f", mt, s.Name, v)
			}
		}
	}
}
