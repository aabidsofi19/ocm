package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ocm/internal/model"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	return &Store{db: db}, nil
}

func (s *Store) Migrate(ctx context.Context) error {
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS services(
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			repository TEXT NOT NULL,
			UNIQUE(name, repository)
		);`,
		`CREATE TABLE IF NOT EXISTS metrics(
			id INTEGER PRIMARY KEY,
			service_id INTEGER NOT NULL,
			metric_type TEXT NOT NULL,
			metric_value REAL NOT NULL,
			timestamp TEXT NOT NULL,
			FOREIGN KEY(service_id) REFERENCES services(id)
		);`,
		`CREATE TABLE IF NOT EXISTS composite_scores(
			id INTEGER PRIMARY KEY,
			service_id INTEGER NOT NULL,
			ocm_score REAL NOT NULL,
			timestamp TEXT NOT NULL,
			FOREIGN KEY(service_id) REFERENCES services(id)
		);`,
		`CREATE TABLE IF NOT EXISTS metric_evidence(
			id INTEGER PRIMARY KEY,
			service_id INTEGER NOT NULL,
			metric_type TEXT NOT NULL,
			component TEXT NOT NULL,
			evidence_key TEXT,
			evidence_value TEXT,
			source_path TEXT NOT NULL DEFAULT '',
			manifest_kind TEXT,
			manifest_name TEXT,
			timestamp TEXT NOT NULL,
			FOREIGN KEY(service_id) REFERENCES services(id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_evidence_service_metric_time ON metric_evidence(service_id, metric_type, timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_service_time ON metrics(service_id, timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_scores_service_time ON composite_scores(service_id, timestamp);`,
	}
	for _, st := range stmts {
		if _, err := s.db.ExecContext(ctx, st); err != nil {
			return err
		}
	}
	return nil
}

type SaveRunInput struct {
	RunAt    time.Time
	Services []model.AnalysisServiceResult
}

func (s *Store) SaveRun(ctx context.Context, in SaveRunInput) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, svc := range in.Services {
		serviceID, err := upsertService(ctx, tx, svc.Name, svc.Repository)
		if err != nil {
			return err
		}
		for mt, val := range svc.Metrics {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO metrics(service_id, metric_type, metric_value, timestamp) VALUES(?,?,?,?)`,
				serviceID, string(mt), val, in.RunAt.UTC().Format(time.RFC3339Nano)); err != nil {
				return err
			}
		}
		for mt, items := range svc.Evidence {
			for _, ev := range items {
				if _, err := tx.ExecContext(ctx,
					`INSERT INTO metric_evidence(service_id, metric_type, component, evidence_key, evidence_value, source_path, manifest_kind, manifest_name, timestamp)
					 VALUES(?,?,?,?,?,?,?,?,?)`,
					serviceID,
					string(mt),
					ev.Component,
					ev.Key,
					ev.Value,
					ev.SourcePath,
					ev.ManifestKind,
					ev.ManifestName,
					in.RunAt.UTC().Format(time.RFC3339Nano),
				); err != nil {
					return err
				}
			}
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO composite_scores(service_id, ocm_score, timestamp) VALUES(?,?,?)`,
			serviceID, svc.Score, in.RunAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) ListEvidenceLatest(ctx context.Context, serviceID int64, metricType model.MetricType) ([]model.EvidenceItem, error) {
	var latestTS sql.NullString
	if err := s.db.QueryRowContext(ctx,
		`SELECT MAX(timestamp) FROM metric_evidence WHERE service_id=? AND metric_type=?`,
		serviceID, string(metricType)).Scan(&latestTS); err != nil {
		return nil, err
	}
	if !latestTS.Valid || latestTS.String == "" {
		return []model.EvidenceItem{}, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT metric_type, component, evidence_key, evidence_value, source_path, manifest_kind, manifest_name
		 FROM metric_evidence
		 WHERE service_id=? AND metric_type=? AND timestamp=?
		 ORDER BY id ASC`,
		serviceID, string(metricType), latestTS.String)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.EvidenceItem
	for rows.Next() {
		var (
			mt, comp      string
			key, val, src sql.NullString
			kind, name    sql.NullString
		)
		if err := rows.Scan(&mt, &comp, &key, &val, &src, &kind, &name); err != nil {
			return nil, err
		}
		out = append(out, model.EvidenceItem{
			MetricType:   model.MetricType(mt),
			Component:    comp,
			Key:          key.String,
			Value:        val.String,
			SourcePath:   src.String,
			ManifestKind: kind.String,
			ManifestName: name.String,
		})
	}
	return out, rows.Err()
}

func upsertService(ctx context.Context, tx *sql.Tx, name, repo string) (int64, error) {
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO services(name, repository) VALUES(?,?) ON CONFLICT(name, repository) DO NOTHING`,
		name, repo); err != nil {
		return 0, err
	}
	var id int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM services WHERE name=? AND repository=?`, name, repo).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Store) ListServices(ctx context.Context) ([]model.Service, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, repository FROM services ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Service
	for rows.Next() {
		var svc model.Service
		if err := rows.Scan(&svc.ID, &svc.Name, &svc.Repository); err != nil {
			return nil, err
		}
		out = append(out, svc)
	}
	return out, rows.Err()
}

func (s *Store) GetMetricSeries(ctx context.Context, serviceID int64, metricType model.MetricType) ([]model.MetricPoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT metric_type, metric_value, timestamp FROM metrics WHERE service_id=? AND metric_type=? ORDER BY timestamp ASC`,
		serviceID, string(metricType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.MetricPoint
	for rows.Next() {
		var (
			mt string
			v  float64
			ts string
		)
		if err := rows.Scan(&mt, &v, &ts); err != nil {
			return nil, err
		}
		pt, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return nil, err
		}
		out = append(out, model.MetricPoint{MetricType: model.MetricType(mt), Value: v, Timestamp: pt})
	}
	return out, rows.Err()
}

func (s *Store) GetScoreSeries(ctx context.Context, serviceID int64) ([]model.ScorePoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT ocm_score, timestamp FROM composite_scores WHERE service_id=? ORDER BY timestamp ASC`,
		serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ScorePoint
	for rows.Next() {
		var (
			score float64
			ts    string
		)
		if err := rows.Scan(&score, &ts); err != nil {
			return nil, err
		}
		pt, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return nil, err
		}
		out = append(out, model.ScorePoint{Score: score, Timestamp: pt})
	}
	return out, rows.Err()
}
