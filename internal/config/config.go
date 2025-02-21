package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	APIKey    string       `json:"api_key"`
	ModelName string       `json:"model_name"`
	BaseURL   string       `json:"base_url"`
	Ding      DingConfig   `json:"ding"`
	Output    OutputConfig `json:"output"`
	Cache     CacheConfig  `json:"cache"`
	Review    ReviewConfig `json:"review"`
}

// DingConfig 钉钉配置
type DingConfig struct {
	Enabled bool   `json:"enabled"`
	Webhook string `json:"webhook"`
	Secret  string `json:"secret"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Dir     string       `json:"dir"`
	Format  []string     `json:"format"`
	Reports ReportConfig `json:"reports"`
}

// ReportConfig 报告配置
type ReportConfig struct {
	IncludeGitInfo bool       `json:"include_git_info"`
	IncludeStats   bool       `json:"include_stats"`
	Template       string     `json:"template"`
	CSSTemplate    string     `json:"css_template"`
	PDFOptions     PDFOptions `json:"pdf_options"`
}

// PDFOptions PDF 导出选项
type PDFOptions struct {
	PageSize         string `json:"page_size"`
	WithLineNumbers  bool   `json:"with_line_numbers"`
	HighlightChanges bool   `json:"highlight_changes"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled    bool   `json:"enabled"`
	Dir        string `json:"dir"`
	ExpireDays int    `json:"expire_days"`
}

// ReviewTemplate 评审模板配置
type ReviewTemplate struct {
	SystemPrompt string   `json:"system_prompt"`
	FocusPoints  []string `json:"focus_points"`
}

// ReviewConfig 评审配置
type ReviewConfig struct {
	Template       string                    `json:"template"`
	Templates      map[string]ReviewTemplate `json:"templates"`
	IgnorePatterns []string                  `json:"ignore_patterns"`
	MaxDiffSize    int                       `json:"max_diff_size"`
}

var globalConfig *Config

func Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	globalConfig = &Config{}
	if err := json.NewDecoder(file).Decode(globalConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	return nil
}

func Get() *Config {
	return globalConfig
}
