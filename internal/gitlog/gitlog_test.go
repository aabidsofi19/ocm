package gitlog

import (
	"testing"
	"time"
)

func TestParseGitLogOutput_MultipleCommits(t *testing.T) {
	output := `COMMIT:abc1234567890 2025-01-15T10:30:00Z

svc-a/deployment.yaml
svc-a/configmap.yaml
svc-b/service.yaml

COMMIT:def9876543210 2025-01-14T08:00:00Z

svc-a/deployment.yaml
`
	commits := parseGitLogOutput(output)
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	c0 := commits[0]
	if c0.Hash != "abc1234567890" {
		t.Fatalf("expected hash abc1234567890, got %s", c0.Hash)
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2025-01-15T10:30:00Z")
	if !c0.Timestamp.Equal(expectedTime) {
		t.Fatalf("expected timestamp %v, got %v", expectedTime, c0.Timestamp)
	}
	if len(c0.Files) != 3 {
		t.Fatalf("expected 3 files in commit 0, got %d", len(c0.Files))
	}

	c1 := commits[1]
	if c1.Hash != "def9876543210" {
		t.Fatalf("expected hash def9876543210, got %s", c1.Hash)
	}
	if len(c1.Files) != 1 {
		t.Fatalf("expected 1 file in commit 1, got %d", len(c1.Files))
	}
}

func TestParseGitLogOutput_EmptyOutput(t *testing.T) {
	commits := parseGitLogOutput("")
	if len(commits) != 0 {
		t.Fatalf("expected 0 commits for empty output, got %d", len(commits))
	}
}

func TestParseGitLogOutput_SingleCommitNoFiles(t *testing.T) {
	output := "COMMIT:aaa1111 2025-06-01T00:00:00Z\n"
	commits := parseGitLogOutput(output)
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
	if commits[0].Hash != "aaa1111" {
		t.Fatalf("expected hash aaa1111, got %s", commits[0].Hash)
	}
	if len(commits[0].Files) != 0 {
		t.Fatalf("expected 0 files, got %d", len(commits[0].Files))
	}
}

func TestParseGitLogOutput_MalformedTimestamp(t *testing.T) {
	output := "COMMIT:bbb2222 not-a-date\nsome/file.yaml\n"
	commits := parseGitLogOutput(output)
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
	if commits[0].Hash != "bbb2222" {
		t.Fatalf("expected hash bbb2222, got %s", commits[0].Hash)
	}
	if commits[0].Timestamp.IsZero() != true {
		t.Fatalf("expected zero timestamp for malformed date, got %v", commits[0].Timestamp)
	}
	if len(commits[0].Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(commits[0].Files))
	}
}

func TestParseGitLogOutput_HashOnly(t *testing.T) {
	// COMMIT line with hash but no timestamp
	output := "COMMIT:ccc3333\nfile.yaml\n"
	commits := parseGitLogOutput(output)
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
	if commits[0].Hash != "ccc3333" {
		t.Fatalf("expected hash ccc3333, got %s", commits[0].Hash)
	}
}

func TestOptionsSetDefaults(t *testing.T) {
	opts := Options{}
	opts.setDefaults()

	if opts.Window != DefaultWindow {
		t.Fatalf("expected default window %v, got %v", DefaultWindow, opts.Window)
	}
	if opts.Now == nil {
		t.Fatal("expected Now to be set")
	}
	if len(opts.ConfigExtensions) == 0 {
		t.Fatal("expected default config extensions")
	}

	// Verify known extensions are present
	extSet := map[string]bool{}
	for _, e := range opts.ConfigExtensions {
		extSet[e] = true
	}
	for _, want := range []string{".yaml", ".yml", ".json"} {
		if !extSet[want] {
			t.Fatalf("expected extension %s in defaults", want)
		}
	}
}
