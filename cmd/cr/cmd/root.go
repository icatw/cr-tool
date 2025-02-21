package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/icatw/cr-tool/pkg/config"
	"github.com/icatw/cr-tool/pkg/exporter"
	"github.com/icatw/cr-tool/pkg/review"
	"github.com/spf13/cobra"
)

var (
	configFile string
	outputDir  string
	format     string
)

var rootCmd = &cobra.Command{
	Use:   "cr",
	Short: "代码评审工具",
	Long: `一个基于 AI 的代码评审工具，支持多种输出格式。
使用示例：
  git diff | cr                    # 使用默认配置评审当前改动
  cr -c config.json               # 指定配置文件
  cr -o ./reports -f html        # 指定输出目录和格式`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 加载配置
		if err := config.Init(); err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 覆盖配置选项
		cfg := config.Get()
		if outputDir != "" {
			cfg.Output.Dir = outputDir
		}
		if format != "" {
			cfg.Output.Format = []string{format}
		}

		// 读取 diff 内容
		var diffContent string
		if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
			// 从管道读取
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("读取输入失败: %w", err)
			}
			diffContent = string(data)
		} else {
			return fmt.Errorf("请通过管道提供 git diff 内容")
		}

		// 执行评审
		reviewer := review.New()
		history, err := reviewer.Review(diffContent)
		if err != nil {
			return fmt.Errorf("代码评审失败: %w", err)
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

			fmt.Printf("评审报告已保存到: %s\n", outputPath)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.json", "配置文件路径")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "输出目录")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "", "输出格式(markdown/html/pdf)")

	config.SetConfigFile(configFile)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
