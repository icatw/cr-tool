package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cr/internal/config"
	"github.com/cr/internal/exporter"
	"github.com/cr/internal/review"
)

func main() {
	configFile := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	if err := config.Load(*configFile); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	reviewer := review.New()

	// 读取 diff 内容
	var diffContent strings.Builder
	if _, err := io.Copy(&diffContent, os.Stdin); err != nil {
		log.Fatalf("读取 diff 内容失败: %v", err)
	}

	// 执行评审
	history, err := reviewer.Review(diffContent.String())
	if err != nil {
		log.Fatalf("代码评审失败: %v", err)
	}

	// 导出结果
	for _, format := range config.Get().Output.Format {
		exp, err := exporter.New(format)
		if err != nil {
			log.Printf("创建导出器失败 (%s): %v", format, err)
			continue
		}

		if _, err := exp.Export(history); err != nil {
			log.Printf("导出失败 (%s): %v", format, err)
		}
	}
}
