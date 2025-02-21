# CR Tool - AI 代码评审工具

一个基于 AI 的代码评审工具，支持多种输出格式，可以自动分析代码变更并生成详细的评审报告。

## 特性

- 🤖 基于 AI 的智能代码评审
- 📊 多种输出格式支持 (Markdown/HTML/PDF)
- 💾 本地缓存支持，避免重复评审
- 🔄 与 Git 深度集成
- 📈 详细的统计分析
- ⚙️ 灵活的配置选项

## 安装

```bash
go install github.com/icatw/cr-tool/cmd/cr@latest
```

## 快速开始

1. 初始化配置：
```bash
cr init
# 根据提示输入 API Key
```

2. 评审当前改动：
```bash
git diff | cr
```

3. 使用特定格式导出：
```bash
git diff | cr -f html
```

## 配置说明

默认配置文件位置：`~/.cr-tool/config.json`

```json
{
  "api_key": "your_api_key",
  "model_name": "qwen-plus",
  "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
  "output": {
    "dir": "./review_results",
    "format": ["markdown"]
  },
  "cache": {
    "enabled": true,
    "dir": "./.cache/code_review",
    "expire_days": 7
  },
  "review": {
    "template": "default",
    "templates": {
      "default": {
        "system_prompt": "你是一个专业的代码评审员...",
        "focus_points": [
          "代码质量",
          "性能优化",
          "安全问题"
        ]
      }
    },
    "ignore_patterns": [
      "*.min.js",
      "vendor/*"
    ],
    "max_diff_size": 2000
  }
}
```

## 命令行选项

```bash
Usage:
  cr [flags]
  cr [command]

Commands:
  init        初始化配置文件
  help        查看帮助信息

Flags:
  -c, --config string   配置文件路径 (默认 "config.json")
  -o, --output string   输出目录
  -f, --format string   输出格式(markdown/html/pdf)
  -h, --help           查看帮助信息
```

## 作为库使用

基础用法：
```go
import "github.com/icatw/cr-tool/pkg/review"

func main() {
    reviewer := review.New()
    history, err := reviewer.Review(diffContent)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(history.ReviewResult)
}
```

自定义配置：
```go
import (
    "github.com/icatw/cr-tool/pkg/config"
    "github.com/icatw/cr-tool/pkg/exporter"
    "github.com/icatw/cr-tool/pkg/review"
)

func main() {
    cfg := &config.Config{
        APIKey:    "your_api_key",
        ModelName: "qwen-plus",
        Output: config.OutputConfig{
            Dir:    "./reports",
            Format: []string{"markdown", "html"},
        },
    }

    reviewer := review.New()
    history, err := reviewer.Review(diffContent)
    if err != nil {
        log.Fatal(err)
    }

    // 导出结果
    for _, format := range cfg.Output.Format {
        exp, err := exporter.New(format)
        if err != nil {
            continue
        }
        outputPath, err := exp.Export(history)
        if err != nil {
            continue
        }
        fmt.Printf("报告已保存到: %s\n", outputPath)
    }
}
```

## 项目结构

```
cr-tool/
├── cmd/
│   └── cr/              # 命令行工具
├── pkg/
│   ├── config/          # 配置管理
│   ├── review/          # 评审核心功能
│   ├── exporter/        # 导出功能
│   └── git/             # Git 相关功能
├── examples/            # 使用示例
├── go.mod
└── README.md
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
