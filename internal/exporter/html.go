package exporter

import (
	"fmt"
	"github.com/your/project/internal/config"
	"github.com/your/project/internal/model"
	"strings"
)

type HTMLExporter struct {
	cssTemplate string
}

func NewHTMLExporter() *HTMLExporter {
	return &HTMLExporter{
		cssTemplate: config.Get().Output.Reports.CSSTemplate,
	}
}

func (e *HTMLExporter) Export(history *model.ReviewHistory) (string, error) {
	var html strings.Builder

	// 添加 HTML 头部和样式
	html.WriteString(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>代码评审报告</title>
	<style>
		%s
	</style>
</head>
<body>`, e.getCSS()))

	// ... HTML 内容生成 ...

	html.WriteString("</body></html>")
	return html.String(), nil
}

func (e *HTMLExporter) getCSS() string {
	// 根据模板返回不同的 CSS
	switch e.cssTemplate {
	case "github":
		return `
			body { font-family: -apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif; }
			.container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
			/* ... 更多 GitHub 风格的 CSS ... */
		`
	default:
		return `
			body { font-family: Arial, sans-serif; }
			.container { margin: 20px; }
		`
	}
}
