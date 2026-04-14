// Package mapper handles .docpush.yaml configuration and frontmatter parsing.
package mapper

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level .docpush.yaml configuration.
type Config struct {
	Targets map[string]TargetEntry    `yaml:"targets"`
	Pages   map[string]PageOverrides `yaml:"pages"`
}

// TargetEntry holds configuration for a specific publish target.
type TargetEntry struct {
	Type     string         `yaml:"type"`
	URL      string         `yaml:"url"`
	Space    string         `yaml:"space"`
	Defaults DefaultsEntry  `yaml:"defaults"`
}

// DefaultsEntry holds default values for a target.
type DefaultsEntry struct {
	ParentID      string   `yaml:"parentId"`
	Labels        []string `yaml:"labels"`
	SourceBaseURL string   `yaml:"sourceBaseUrl"`
	ReadOnly      bool     `yaml:"readOnly"`
}

// PageOverrides holds per-page, per-target overrides.
type PageOverrides map[string]map[string]interface{}

// LoadConfig reads and parses a .docpush.yaml file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	return ParseConfig(data)
}

// ParseConfig parses .docpush.yaml content.
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// GetPageMeta returns the merged metadata for a file path and target.
// Frontmatter overrides take precedence over the config file.
func GetPageMeta(cfg *Config, filePath, target string, frontmatter map[string]interface{}) map[string]any {
	meta := make(map[string]any)

	// Start with target defaults
	if t, ok := cfg.Targets[target]; ok {
		if t.Defaults.ParentID != "" {
			meta["parentId"] = t.Defaults.ParentID
		}
	}

	// Layer on page-specific config
	if pageOverrides, ok := cfg.Pages[filePath]; ok {
		if targetOverrides, ok := pageOverrides[target]; ok {
			for k, v := range targetOverrides {
				if v != nil {
					meta[k] = v
				}
			}
		}
	}

	// Layer on frontmatter overrides (highest priority)
	if frontmatter != nil {
		// Pull top-level title from frontmatter
		if t, ok := frontmatter["title"]; ok {
			if s, ok := t.(string); ok && s != "" {
				meta["title"] = s
			}
		}

		if dp, ok := frontmatter["docpush"]; ok {
			if dpMap, ok := dp.(map[string]interface{}); ok {
				if targetFM, ok := dpMap[target]; ok {
					if targetMap, ok := targetFM.(map[string]interface{}); ok {
						for k, v := range targetMap {
							if v != nil {
								meta[k] = v
							}
						}
					}
				}
			}
			// Also handle map[interface{}]interface{} from some YAML parsers
			if dpMap, ok := dp.(map[interface{}]interface{}); ok {
				if targetFM, ok := dpMap[target]; ok {
					if targetMap, ok := targetFM.(map[interface{}]interface{}); ok {
						for k, v := range targetMap {
							if v != nil {
								meta[fmt.Sprintf("%v", k)] = v
							}
						}
					}
				}
			}
		}
	}

	return meta
}

// WritePageID updates the pageId for a file in .docpush.yaml.
func WritePageID(configPath, filePath, target, pageID string) error {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	if cfg.Pages == nil {
		cfg.Pages = make(map[string]PageOverrides)
	}
	if cfg.Pages[filePath] == nil {
		cfg.Pages[filePath] = make(PageOverrides)
	}
	if cfg.Pages[filePath][target] == nil {
		cfg.Pages[filePath][target] = make(map[string]interface{})
	}
	cfg.Pages[filePath][target]["pageId"] = pageID

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(configPath, data, 0644)
}

// WriteConfig serializes and writes the full config to the given path.
func WriteConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
