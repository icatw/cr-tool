package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	content := `{
		"api_key": "test_key",
		"model_name": "test_model",
		"base_url": "http://test.api"
	}`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// 设置配置文件路径
	SetConfigFile(configFile)

	// 测试初始化
	if err := Init(); err != nil {
		t.Fatal(err)
	}

	// 验证配置
	cfg := Get()
	if cfg.APIKey != "test_key" {
		t.Errorf("expected APIKey to be 'test_key', got '%s'", cfg.APIKey)
	}
}
