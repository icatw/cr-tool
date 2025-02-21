package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/icatw/cr-tool/pkg/config"
	"github.com/icatw/cr-tool/pkg/review"
)

type MarkdownExporter struct {
	config *config.Config
}

func NewMarkdownExporter() *MarkdownExporter {
	return &MarkdownExporter{
		config: config.Get(),
	}
}

func (e *MarkdownExporter) Export(history *review.ReviewHistory) (string, error) {
	var md strings.Builder

	// 添加标题
	md.WriteString("# 代码评审报告\n\n")

	// Git 信息
	if history.GitInfo != nil {
		md.WriteString("## Git 信息\n\n")
		md.WriteString(fmt.Sprintf("- 分支: `%s`\n", history.GitInfo.Branch))
		md.WriteString(fmt.Sprintf("- 提交: `%s`\n", history.GitInfo.CommitHash))
		md.WriteString(fmt.Sprintf("- 作者: %s\n", history.GitInfo.Author))
		md.WriteString(fmt.Sprintf("- 提交信息: %s\n\n", history.GitInfo.CommitMessage))
	}

	// 统计信息
	if history.ReviewStats != nil {
		md.WriteString("## 变更统计\n\n")
		md.WriteString(fmt.Sprintf("- 变更文件数: %d\n", history.ReviewStats.FilesChanged))
		md.WriteString(fmt.Sprintf("- 新增行数: %d\n", history.ReviewStats.LinesAdded))
		md.WriteString(fmt.Sprintf("- 删除行数: %d\n\n", history.ReviewStats.LinesDeleted))

		if len(history.ReviewStats.IssuesByLevel) > 0 {
			md.WriteString("### 问题级别统计\n\n")
			for level, count := range history.ReviewStats.IssuesByLevel {
				md.WriteString(fmt.Sprintf("- %s: %d\n", level, count))
			}
			md.WriteString("\n")
		}
	}

	// 评审结果
	md.WriteString("## 评审详情\n\n")
	md.WriteString(history.ReviewResult)

	// 保存文件
	outputDir := e.config.Output.Dir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	filename := fmt.Sprintf("%s_review.md", time.Now().Format("20060102_150405"))
	outputPath := filepath.Join(outputDir, filename)

	if err := os.WriteFile(outputPath, []byte(md.String()), 0644); err != nil {
		return "", fmt.Errorf("保存评审报告失败: %w", err)
	}

	return outputPath, nil
}
