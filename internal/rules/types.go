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
}
