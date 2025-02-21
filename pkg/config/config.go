package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	defaultConfig *Config
	configFile    string
)

// Init 初始化配置
func Init() error {
	v := viper.New()

	// 设置环境变量前缀
	v.SetEnvPrefix("CR_TOOL")
	v.AutomaticEnv()

	// 设置默认值
	setDefaults(v)

	// 按优先级加载配置
	if err := loadConfig(v); err != nil {
		return err
	}

	// 解析配置
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	defaultConfig = &config
	return nil
}

// Get 获取配置
func Get() *Config {
	return defaultConfig
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	v.SetDefault("model_name", "qwen-plus")
	v.SetDefault("base_url", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions")
	v.SetDefault("output.dir", "./review_results")
	v.SetDefault("output.format", []string{"markdown"})
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.dir", "./.cache/code_review")
	v.SetDefault("cache.expire_days", 7)
	v.SetDefault("review.template", "default")
	v.SetDefault("review.max_diff_size", 2000)
}

// loadConfig 加载配置文件
func loadConfig(v *viper.Viper) error {
	// 1. 检查命令行指定的配置文件
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
		return nil
	}

	// 2. 检查项目目录
	v.SetConfigName(".cr-tool")
	v.SetConfigType("json")
	v.AddConfigPath(".")

	// 3. 检查用户目录
	home, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(home, ".cr-tool"))
	}

	// 4. 检查系统目录
	v.AddConfigPath("/etc/cr-tool")

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	return nil
}

// SetConfigFile 设置配置文件路径
func SetConfigFile(path string) {
	configFile = path
}
