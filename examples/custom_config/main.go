package main

import (
	"fmt"
	"log"

	"github.com/icatw/cr-tool/pkg/config"
	"github.com/icatw/cr-tool/pkg/exporter"
	"github.com/icatw/cr-tool/pkg/review"
)

func main() {
	// 自定义配置
	cfg := &config.Config{
		APIKey:    "your_api_key",
		ModelName: "qwen-plus",
		BaseURL:   "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		Output: config.OutputConfig{
			Dir:    "./reports",
			Format: []string{"markdown", "html"},
		},
		Cache: config.CacheConfig{
			Enabled:    true,
			Dir:        "./.cache",
			ExpireDays: 7,
		},
	}

	// 创建评审器
	reviewer := review.New()

	// 评审代码
	diffContent := `diff --git a/main.go b/main.go
+ func main() {
+    // TODO: 实现主函数
+ }
`
	history, err := reviewer.Review(diffContent)
	if err != nil {
		log.Fatalf("评审失败: %v", err)
	}

	// 导出结果
	for _, format := range cfg.Output.Format {
		exp, err := exporter.New(format)
		if err != nil {
			log.Printf("创建导出器失败 (%s): %v", format, err)
			continue
		}

		outputPath, err := exp.Export(history)
		if err != nil {
			log.Printf("导出失败 (%s): %v", format, err)
			continue
		}

		fmt.Printf("报告已保存到: %s\n", outputPath)
	}
}
