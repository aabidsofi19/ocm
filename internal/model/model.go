package model

import "time"

type MetricType string

const (
	MetricCSA MetricType = "CSA" // Configuration Surface Area
	MetricDD  MetricType = "DD"  // Dependency Depth
	MetricDB  MetricType = "DB"  // Dependency Breadth
	MetricCV  MetricType = "CV"  // Change Volatility
	MetricFE  MetricType = "FE"  // Failure Exposure
	MetricCDR MetricType = "CDR" // Configuration Drift Risk
)

// AllMetrics lists every metric type in canonical order.
var AllMetrics = []MetricType{MetricCSA, MetricDD, MetricDB, MetricCV, MetricFE, MetricCDR}

// DefaultWeights defines the default weight for each metric in the composite
// OCM score. The weights are equal (1/6 each). They MUST sum to 1.0.
// Weights are configurable per the composite scoring spec.
var DefaultWeights = map[MetricType]float64{
	MetricCSA: 1.0 / 6.0,
	MetricDD:  1.0 / 6.0,
	MetricDB:  1.0 / 6.0,
	MetricCV:  1.0 / 6.0,
	MetricFE:  1.0 / 6.0,
	MetricCDR: 1.0 / 6.0,
}

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
	Name         string                        `json:"name"`
	Repository   string                        `json:"repository"`
	Dependencies []string                      `json:"dependencies,omitempty"`
	Metrics      map[MetricType]float64        `json:"metrics"`
	Normalized   map[MetricType]float64        `json:"normalized"`
	Score        float64                       `json:"score"`
	Evidence     map[MetricType][]EvidenceItem `json:"evidence,omitempty"`
}

type EvidenceItem struct {
	MetricType   MetricType `json:"metricType"`
	Component    string     `json:"component"` // env|port|resource|replica|spec_key
	Key          string     `json:"key,omitempty"`
	Value        string     `json:"value,omitempty"`
	SourcePath   string     `json:"sourcePath"`
	ManifestKind string     `json:"manifestKind,omitempty"`
	ManifestName string     `json:"manifestName,omitempty"`
}

type AnalysisResult struct {
	RunAt    time.Time               `json:"runAt"`
	RepoPath string                  `json:"repoPath"`
	Services []AnalysisServiceResult `json:"services"`
	Warnings []string                `json:"warnings,omitempty"`
}
