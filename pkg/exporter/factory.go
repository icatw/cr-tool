package exporter

import "fmt"

// New 创建导出器
func New(format string) (Exporter, error) {
	switch Format(format) {
	case FormatMarkdown:
		return NewMarkdownExporter(), nil
	case FormatHTML:
		return NewHTMLExporter(), nil
	case FormatPDF:
		return NewPDFExporter(), nil
	default:
		return nil, fmt.Errorf("不支持的导出格式: %s", format)
	}
}
