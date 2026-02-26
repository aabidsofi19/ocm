package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"ocm/internal/model"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *Store {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return store
}

func TestSaveAndRetrieveEvidence_AllMetrics(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()
	runAt := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	services := []model.AnalysisServiceResult{
		{
			Name:       "api",
			Repository: "/tmp/repo",
			Metrics: map[model.MetricType]float64{
				model.MetricCSA: 10,
				model.MetricDD:  2,
				model.MetricDB:  3,
				model.MetricCV:  5,
				model.MetricFE:  1,
				model.MetricCDR: 1,
			},
			Score: 0.42,
			Evidence: map[model.MetricType][]model.EvidenceItem{
				model.MetricCSA: {
					{MetricType: model.MetricCSA, Component: "env", Key: "PORT", Value: "8080", SourcePath: "api/deploy.yaml", ManifestKind: "Deployment", ManifestName: "api"},
				},
				model.MetricDD: {
					{MetricType: model.MetricDD, Component: "dep_chain", Key: "depth=2", Value: "api -> cache -> db"},
					{MetricType: model.MetricDD, Component: "direct_dep", Key: "cache", Value: "api depends on cache"},
				},
				model.MetricDB: {
					{MetricType: model.MetricDB, Component: "downstream_dep", Key: "cache", Value: "api -> cache"},
				},
				model.MetricCV: {
					{MetricType: model.MetricCV, Component: "commit", Key: "abc12345", Value: "commit abc12345"},
				},
				model.MetricFE: {
					{MetricType: model.MetricFE, Component: "loadbalancer", Key: "api-svc", Value: "LoadBalancer service", SourcePath: "api/svc.yaml", ManifestKind: "Service", ManifestName: "api-svc"},
				},
				model.MetricCDR: {
					{MetricType: model.MetricCDR, Component: "env_override", Key: "config.yaml", Value: "found in 2 envs: dev, prod"},
				},
			},
		},
	}

	if err := store.SaveRun(ctx, SaveRunInput{RunAt: runAt, Services: services}); err != nil {
		t.Fatalf("SaveRun: %v", err)
	}

	// Retrieve service ID.
	svcs, err := store.ListServices(ctx)
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	if len(svcs) != 1 {
		t.Fatalf("expected 1 service, got %d", len(svcs))
	}
	svcID := svcs[0].ID

	// Verify evidence for each metric type.
	for _, mt := range model.AllMetrics {
		evs, err := store.ListEvidenceLatest(ctx, svcID, mt)
		if err != nil {
			t.Fatalf("ListEvidenceLatest(%s): %v", mt, err)
		}
		expected := services[0].Evidence[mt]
		if len(evs) != len(expected) {
			t.Errorf("metric %s: expected %d evidence items, got %d", mt, len(expected), len(evs))
			continue
		}
		for i, ev := range evs {
			if ev.Component != expected[i].Component {
				t.Errorf("metric %s item %d: component = %q, want %q", mt, i, ev.Component, expected[i].Component)
			}
			if ev.Key != expected[i].Key {
				t.Errorf("metric %s item %d: key = %q, want %q", mt, i, ev.Key, expected[i].Key)
			}
			if ev.Value != expected[i].Value {
				t.Errorf("metric %s item %d: value = %q, want %q", mt, i, ev.Value, expected[i].Value)
			}
		}
	}
}

func TestSaveRun_EmptySourcePath(t *testing.T) {
	// Verify that evidence items with empty SourcePath are saved successfully.
	// This is important for pipeline-computed evidence (DD, DB, CV, CDR).
	store := setupTestDB(t)
	ctx := context.Background()
	runAt := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	services := []model.AnalysisServiceResult{
		{
			Name:       "svc",
			Repository: "/tmp/repo",
			Metrics: map[model.MetricType]float64{
				model.MetricDD: 1,
			},
			Score: 0.1,
			Evidence: map[model.MetricType][]model.EvidenceItem{
				model.MetricDD: {
					{MetricType: model.MetricDD, Component: "dep_chain", Key: "depth=1", Value: "svc -> other", SourcePath: ""},
				},
			},
		},
	}

	if err := store.SaveRun(ctx, SaveRunInput{RunAt: runAt, Services: services}); err != nil {
		t.Fatalf("SaveRun with empty source_path should succeed: %v", err)
	}

	svcs, err := store.ListServices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	evs, err := store.ListEvidenceLatest(ctx, svcs[0].ID, model.MetricDD)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("expected 1 evidence item, got %d", len(evs))
	}
	if evs[0].Value != "svc -> other" {
		t.Errorf("evidence value = %q, want %q", evs[0].Value, "svc -> other")
	}
}
