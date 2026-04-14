package confluence

// TargetConfig holds Confluence-specific configuration from .docpush.yaml.
type TargetConfig struct {
	Type     string         `yaml:"type"`
	URL      string         `yaml:"url"`
	Space    string         `yaml:"space"`
	Defaults DefaultsConfig `yaml:"defaults"`
}

// DefaultsConfig holds default values for page creation.
type DefaultsConfig struct {
	ParentID      string   `yaml:"parentId"`
	Labels        []string `yaml:"labels"`
	SourceBaseURL string   `yaml:"sourceBaseUrl"`
	ReadOnly      bool     `yaml:"readOnly"`
}

// PageConfig holds per-page Confluence-specific configuration.
type PageConfig struct {
	PageID   string `yaml:"pageId"`
	ParentID string `yaml:"parentId"`
	Title    string `yaml:"title"`
}
