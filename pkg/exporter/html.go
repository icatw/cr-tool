package exporter

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/icatw/cr-tool/pkg/config"
	"github.com/icatw/cr-tool/pkg/review"
)

type HTMLExporter struct {
	config *config.Config
}

func NewHTMLExporter() *HTMLExporter {
	return &HTMLExporter{
		config: config.Get(),
	}
}

func (e *HTMLExporter) Export(history *review.ReviewHistory) (string, error) {
	var b strings.Builder

	// 添加 HTML 头部和样式
	b.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>代码评审报告</title>
    <style>%s</style>
</head>
<body>
<div class="container">`, e.getCSS()))

	// 标题
	b.WriteString(`<h1>代码评审报告</h1>`)

	// Git 信息
	if history.GitInfo != nil {
		b.WriteString(`<div class="git-info">
			<h2>Git 信息</h2>
			<table>
				<tr><td>分支：</td><td><code>`)
		b.WriteString(template.HTMLEscapeString(history.GitInfo.Branch))
		b.WriteString(`</code></td></tr>
				<tr><td>提交：</td><td><code>`)
		b.WriteString(template.HTMLEscapeString(history.GitInfo.CommitHash))
		b.WriteString(`</code></td></tr>
				<tr><td>作者：</td><td>`)
		b.WriteString(template.HTMLEscapeString(history.GitInfo.Author))
		b.WriteString(`</td></tr>
				<tr><td>提交信息：</td><td>`)
		b.WriteString(template.HTMLEscapeString(history.GitInfo.CommitMessage))
		b.WriteString(`</td></tr>
			</table>
		</div>`)
	}

	// 统计信息
	if history.ReviewStats != nil {
		b.WriteString(`<div class="stats">
			<h2>变更统计</h2>
			<div class="stats-grid">
				<div class="stat-item">
					<div class="stat-value">`)
		b.WriteString(fmt.Sprintf("%d", history.ReviewStats.FilesChanged))
		b.WriteString(`</div>
					<div class="stat-label">变更文件数</div>
				</div>
				<div class="stat-item">
					<div class="stat-value">`)
		b.WriteString(fmt.Sprintf("%d", history.ReviewStats.LinesAdded))
		b.WriteString(`</div>
					<div class="stat-label">新增行数</div>
				</div>
				<div class="stat-item">
					<div class="stat-value">`)
		b.WriteString(fmt.Sprintf("%d", history.ReviewStats.LinesDeleted))
		b.WriteString(`</div>
					<div class="stat-label">删除行数</div>
				</div>
			</div>`)

		// 问题级别统计
		if len(history.ReviewStats.IssuesByLevel) > 0 {
			b.WriteString(`<h3>问题级别统计</h3>
			<div class="issues-by-level">`)
			for level, count := range history.ReviewStats.IssuesByLevel {
				b.WriteString(fmt.Sprintf(`<div class="issue-level %s">
					<span class="level-name">%s</span>
					<span class="level-count">%d</span>
				</div>`, strings.ToLower(level), level, count))
			}
			b.WriteString(`</div>`)
		}
		b.WriteString(`</div>`)
	}

	// 评审结果
	b.WriteString(`<div class="review-result">
		<h2>评审详情</h2>
		<div class="markdown-body">`)
	b.WriteString(formatMarkdown(history.ReviewResult))
	b.WriteString(`</div>
	</div>`)

	// 添加页脚
	b.WriteString(fmt.Sprintf(`
		<div class="footer">
			生成时间：%s
		</div>
	</div></body></html>`, time.Now().Format("2006-01-02 15:04:05")))

	// 保存文件
	outputDir := e.config.Output.Dir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	filename := fmt.Sprintf("%s_review.html", time.Now().Format("20060102_150405"))
	outputPath := filepath.Join(outputDir, filename)

	if err := os.WriteFile(outputPath, []byte(b.String()), 0644); err != nil {
		return "", fmt.Errorf("保存评审报告失败: %w", err)
	}

	return outputPath, nil
}

// getCSS 获取 CSS 样式
func (e *HTMLExporter) getCSS() string {
	return `
		:root {
			--bg-color: #f6f8fa;
			--border-color: #d0d7de;
		}
		body { 
			font-family: -apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif; 
			line-height: 1.5;
			color: #24292f;
			margin: 0;
			padding: 20px;
		}
		.container { 
			max-width: 1200px; 
			margin: 0 auto; 
			background: white;
			padding: 2rem;
			border-radius: 6px;
			box-shadow: 0 1px 3px rgba(0,0,0,0.12);
		}
		h1, h2, h3 { margin-top: 1.5em; margin-bottom: 1em; }
		h1 { padding-bottom: .3em; border-bottom: 1px solid var(--border-color); }
		.git-info table { border-collapse: collapse; }
		.git-info td { padding: .5em 1em .5em 0; }
		.stats-grid {
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
			gap: 1rem;
			margin: 1rem 0;
		}
		.stat-item {
			background: var(--bg-color);
			border: 1px solid var(--border-color);
			border-radius: 6px;
			padding: 1rem;
			text-align: center;
		}
		.stat-value { font-size: 2rem; font-weight: bold; }
		.stat-label { color: #666; margin-top: 0.5rem; }
		.issues-by-level {
			display: flex;
			gap: 1rem;
			margin: 1rem 0;
		}
		.issue-level {
			padding: 0.5rem 1rem;
			border-radius: 6px;
			display: flex;
			gap: 0.5rem;
			align-items: center;
		}
		.issue-level.严重 { background: #ffebe9; color: #cf222e; }
		.issue-level.中等 { background: #fff8c5; color: #9a6700; }
		.issue-level.低 { background: #ddf4ff; color: #0969da; }
		.review-result { margin-top: 2rem; }
		.markdown-body {
			background: white;
			padding: 1rem;
			border: 1px solid var(--border-color);
			border-radius: 6px;
		}
		.footer {
			margin-top: 2rem;
			padding-top: 1rem;
			border-top: 1px solid var(--border-color);
			color: #666;
			font-size: 0.9rem;
		}
		code { 
			background: var(--bg-color);
			padding: 0.2em 0.4em;
			border-radius: 3px;
			font-size: 85%;
			font-family: SFMono-Regular,Consolas,Liberation Mono,Menlo,monospace;
		}
	`
}

// formatMarkdown 简单的 Markdown 转 HTML
func formatMarkdown(md string) string {
	lines := strings.Split(md, "\n")
	var b strings.Builder
	inList := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if inList {
				b.WriteString("</ul>\n")
				inList = false
			}
			b.WriteString("<p></p>\n")
			continue
		}

		switch {
		case strings.HasPrefix(line, "# "):
			b.WriteString("<h1>" + template.HTMLEscapeString(line[2:]) + "</h1>\n")
		case strings.HasPrefix(line, "## "):
			b.WriteString("<h2>" + template.HTMLEscapeString(line[3:]) + "</h2>\n")
		case strings.HasPrefix(line, "### "):
			b.WriteString("<h3>" + template.HTMLEscapeString(line[4:]) + "</h3>\n")
		case strings.HasPrefix(line, "- "):
			if !inList {
				b.WriteString("<ul>\n")
				inList = true
			}
			b.WriteString("<li>" + template.HTMLEscapeString(line[2:]) + "</li>\n")
		default:
			if inList {
				b.WriteString("</ul>\n")
				inList = false
			}
			b.WriteString("<p>" + template.HTMLEscapeString(line) + "</p>\n")
		}
	}

	if inList {
		b.WriteString("</ul>\n")
	}

	return b.String()
}
