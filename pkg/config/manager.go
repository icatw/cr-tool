package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InitConfig 初始化用户配置
func InitConfig(apiKey string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}

	configDir := filepath.Join(home, ".cr-tool")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	config := Config{
		APIKey:    apiKey,
		ModelName: "qwen-plus",
		BaseURL:   "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		Output: OutputConfig{
			Dir:    "./review_results",
			Format: []string{"markdown"},
		},
		Cache: CacheConfig{
			Enabled:    true,
			Dir:        "./.cache/code_review",
			ExpireDays: 7,
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	configFile := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}
