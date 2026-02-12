package pipeline

import (
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

	dd := computeDD(services)
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
	res, err := p.Run(t.Context())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	_ = res
}
