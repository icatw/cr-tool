package exporter

import (
	"fmt"
	"github.com/your/project/internal/model"
	"strings"
)

type MarkdownExporter struct{}

func NewMarkdownExporter() *MarkdownExporter {
	return &MarkdownExporter{}
}

func (e *MarkdownExporter) Export(history *model.ReviewHistory) (string, error) {
	var report strings.Builder

	// 添加标题
	report.WriteString("# 代码评审报告\n\n")

	// Git 信息
	if history.GitInfo != nil {
		report.WriteString("## Git 信息\n\n")
		report.WriteString(fmt.Sprintf("- 分支: `%s`\n", history.GitInfo.Branch))
		report.WriteString(fmt.Sprintf("- 提交: `%s`\n", history.GitInfo.CommitHash))
		report.WriteString(fmt.Sprintf("- 作者: %s\n", history.GitInfo.Author))
		report.WriteString(fmt.Sprintf("- 提交信息: %s\n\n", history.GitInfo.CommitMessage))
	}

	// 统计信息
	if history.ReviewStats != nil {
		report.WriteString("## 变更统计\n\n")
		report.WriteString(fmt.Sprintf("- 变更文件数: %d\n", history.ReviewStats.FilesChanged))
		report.WriteString(fmt.Sprintf("- 新增行数: %d\n", history.ReviewStats.LinesAdded))
		report.WriteString(fmt.Sprintf("- 删除行数: %d\n\n", history.ReviewStats.LinesDeleted))
	}

	// 评审结果
	report.WriteString("## 评审详情\n\n")
	report.WriteString(history.ReviewResult)

	return report.String(), nil
}
