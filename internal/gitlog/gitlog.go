// Package gitlog extracts commit metadata from a Git repository by shelling
// out to `git log`. This is used by the Change Volatility (CV) metric.
//
// Per spec (20-parser-normalization.md):
//   - Git history MAY be read by shelling out to `git` when available.
//   - Behavior MUST be documented and deterministic.
//
// Implementation notes:
//   - Uses `git log --name-only` with a configurable time window (default: 30 days).
//   - Maps changed file paths to services using the same directory-based heuristic
//     as the main YAML parser (first path component = service name).
//   - Returns per-service commit counts within the time window.
//   - If `git` is not available or the repo has no history, returns empty results
//     with no error (best-effort).
package gitlog

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DefaultWindow is the default time window for CV computation (30 days).
const DefaultWindow = 30 * 24 * time.Hour

// CommitInfo holds the minimal metadata extracted from a single commit.
type CommitInfo struct {
	Hash      string
	Timestamp time.Time
	Files     []string // paths relative to repo root
}

// ServiceChangeFacts holds per-service CV results.
type ServiceChangeFacts struct {
	// CommitCount is the number of unique commits affecting this service's
	// configuration files within the time window.
	CommitCount int
	// Commits lists the commit hashes that touched this service.
	Commits []string
}

// Options configures Git history extraction.
type Options struct {
	// RepoPath is the absolute path to the repository root.
	RepoPath string
	// Window is the lookback duration from Now. Commits older than
	// Now-Window are excluded. Default: 30 days.
	Window time.Duration
	// Now is the reference time. Default: time.Now().
	Now func() time.Time
	// ConfigExtensions lists file extensions considered as "configuration".
	// Default: [".yaml", ".yml", ".json", ".toml", ".env", ".properties", ".conf"]
	ConfigExtensions []string
}

func (o *Options) setDefaults() {
	if o.Window == 0 {
		o.Window = DefaultWindow
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if len(o.ConfigExtensions) == 0 {
		o.ConfigExtensions = []string{
			".yaml", ".yml", ".json", ".toml",
			".env", ".properties", ".conf",
		}
	}
}

// ExtractChangeFacts shells out to `git log` and returns per-service commit
// counts for configuration files changed within the configured time window.
//
// The returned map is keyed by service name (first path component of changed
// files, matching the "dir" service-key strategy).
//
// If git is not installed or the directory is not a git repo, returns (nil, nil).
func ExtractChangeFacts(opts Options) (map[string]*ServiceChangeFacts, error) {
	opts.setDefaults()

	// Check that git is available.
	if _, err := exec.LookPath("git"); err != nil {
		return nil, nil // git not available; best-effort
	}

	now := opts.Now()
	since := now.Add(-opts.Window).UTC().Format(time.RFC3339)

	// git log --since=<date> --name-only --format="COMMIT:%H %aI" -- .
	// This gives us commit boundaries interleaved with file lists.
	cmd := exec.Command("git", "log",
		"--since="+since,
		"--name-only",
		"--format=COMMIT:%H %aI",
		"--",
		".",
	)
	cmd.Dir = opts.RepoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If the repo has no commits or is not a git repo, treat as empty.
		if strings.Contains(stderr.String(), "not a git repository") ||
			strings.Contains(stderr.String(), "does not have any commits") {
			return nil, nil
		}
		return nil, fmt.Errorf("git log: %w: %s", err, stderr.String())
	}

	commits := parseGitLogOutput(stdout.String())

	// Build extension set for fast lookup.
	extSet := map[string]bool{}
	for _, ext := range opts.ConfigExtensions {
		extSet[strings.ToLower(ext)] = true
	}

	// Map commits to services.
	result := map[string]*ServiceChangeFacts{}
	for _, c := range commits {
		// Track which services this commit touches (deduplicate per commit).
		touched := map[string]bool{}
		for _, f := range c.Files {
			ext := strings.ToLower(filepath.Ext(f))
			if !extSet[ext] {
				continue
			}
			parts := strings.SplitN(f, "/", 2)
			if len(parts) == 0 || parts[0] == "" {
				continue
			}
			svcName := parts[0]
			touched[svcName] = true
		}
		for svc := range touched {
			if result[svc] == nil {
				result[svc] = &ServiceChangeFacts{}
			}
			result[svc].CommitCount++
			result[svc].Commits = append(result[svc].Commits, c.Hash)
		}
	}

	return result, nil
}

// parseGitLogOutput parses the output of git log --name-only --format="COMMIT:%H %aI".
func parseGitLogOutput(output string) []CommitInfo {
	var commits []CommitInfo
	var current *CommitInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "COMMIT:") {
			// Save previous commit.
			if current != nil {
				commits = append(commits, *current)
			}
			current = &CommitInfo{}
			rest := strings.TrimPrefix(line, "COMMIT:")
			parts := strings.SplitN(rest, " ", 2)
			if len(parts) >= 1 {
				current.Hash = parts[0]
			}
			if len(parts) >= 2 {
				if t, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[1])); err == nil {
					current.Timestamp = t
				}
			}
			continue
		}
		// File path line.
		if current != nil {
			current.Files = append(current.Files, line)
		}
	}
	// Don't forget the last commit.
	if current != nil {
		commits = append(commits, *current)
	}

	return commits
}
