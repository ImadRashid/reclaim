package rules

type Catalog struct {
	Categories []Category `yaml:"categories"`
	Rules      []Rule     `yaml:"rules"`
}

type Category struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type Rule struct {
	ID              string   `yaml:"id"`
	Category        string   `yaml:"category"`
	Paths           []string `yaml:"paths"`
	ScanFor         string   `yaml:"scan_for"`
	ScanRoot        string   `yaml:"scan_root"`
	ScanMaxDepth    int      `yaml:"scan_max_depth"`
	PathContains    string   `yaml:"path_contains"`
	RequiresSibling string   `yaml:"requires_sibling"`
	Safety          string   `yaml:"safety"`
	Regenerates     bool     `yaml:"regenerates"`
	Cmd             string   `yaml:"cmd"`
	Description     string   `yaml:"description"`
	Detect          string   `yaml:"detect"`
	ActivityCheck   bool     `yaml:"activity_check"`
	ProcessCheck    string   `yaml:"process_check"`

	// v0.2 extensions:

	// DetectCommand gates rule visibility on a probe command exiting 0.
	// Used when Detect == "command", e.g. `docker info` to verify the
	// daemon is running before showing docker-system-prune.
	DetectCommand string `yaml:"detect_command"`

	// Action describes how the rule should be applied. Default behavior
	// (action == "" or "remove") deletes the rule's matched paths via
	// engine.Apply. Setting `action: shell` runs ActionCommand instead.
	Action string `yaml:"action"`

	// ActionCommand is the shell command to run when Action == "shell".
	// stdout/stderr is captured and surfaced to the user.
	ActionCommand string `yaml:"action_command"`

	// WarnIfRunning is a list of process names; if any are running when
	// the rule is being applied, the engine skips the deletion and
	// records a "skipped: <name> is running" result.
	WarnIfRunning []string `yaml:"warn_if_running"`
}
