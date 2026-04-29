package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LogEntry is a single deletion record.
type LogEntry struct {
	Time    time.Time `json:"time"`
	RuleID  string    `json:"rule"`
	Path    string    `json:"path"`
	Freed   int64     `json:"freed_bytes"`
	Failed  bool      `json:"failed,omitempty"`
	Error   string    `json:"error,omitempty"`
	Skipped bool      `json:"skipped,omitempty"`
	Reason  string    `json:"reason,omitempty"`
}

// LogRun appends results to a JSONL file at ~/.reclaim/logs/YYYY-MM-DD.jsonl.
// Returns the path that was written to (or empty + error).
func LogRun(results []Result) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".reclaim", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, time.Now().Format("2006-01-02")+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, r := range results {
		entry := LogEntry{
			Time:    time.Now(),
			RuleID:  r.RuleID,
			Path:    r.Path,
			Freed:   r.Freed,
			Skipped: r.Skipped,
			Reason:  r.Reason,
		}
		if r.Err != nil {
			entry.Failed = true
			entry.Error = r.Err.Error()
		}
		if err := enc.Encode(entry); err != nil {
			return path, err
		}
	}
	return path, nil
}
