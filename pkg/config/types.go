package config

// Config 配置结构
type Config struct {
	APIKey    string       `mapstructure:"api_key"`
	ModelName string       `mapstructure:"model_name"`
	BaseURL   string       `mapstructure:"base_url"`
	Output    OutputConfig `mapstructure:"output"`
	Cache     CacheConfig  `mapstructure:"cache"`
	Review    ReviewConfig `mapstructure:"review"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Dir    string   `mapstructure:"dir"`
	Format []string `mapstructure:"format"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Dir        string `mapstructure:"dir"`
	ExpireDays int    `mapstructure:"expire_days"`
}

// ReviewConfig 评审配置
type ReviewConfig struct {
	Template       string                    `mapstructure:"template"`
	Templates      map[string]ReviewTemplate `mapstructure:"templates"`
	IgnorePatterns []string                  `mapstructure:"ignore_patterns"`
	MaxDiffSize    int                       `mapstructure:"max_diff_size"`
}

// ReviewTemplate 评审模板
type ReviewTemplate struct {
	SystemPrompt string   `mapstructure:"system_prompt"`
	FocusPoints  []string `mapstructure:"focus_points"`
}
