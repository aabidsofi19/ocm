package model

import "time"

type MetricType string

const (
	MetricCSA MetricType = "CSA"
	MetricDD  MetricType = "DD"
)

type Service struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Repository string `json:"repository"`
}

type MetricPoint struct {
	MetricType MetricType `json:"metricType"`
	Value      float64    `json:"value"`
	Timestamp  time.Time  `json:"timestamp"`
}

type ScorePoint struct {
	Score     float64   `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

type AnalysisServiceResult struct {
	Name         string                 `json:"name"`
	Repository   string                 `json:"repository"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Metrics      map[MetricType]float64 `json:"metrics"`
	Normalized   map[MetricType]float64 `json:"normalized"`
	Score        float64                `json:"score"`
}

type AnalysisResult struct {
	RunAt    time.Time               `json:"runAt"`
	RepoPath string                  `json:"repoPath"`
	Services []AnalysisServiceResult `json:"services"`
	Warnings []string                `json:"warnings,omitempty"`
}
