package exporter

import "github.com/icatw/cr-tool/pkg/review"

// Exporter 导出器接口
type Exporter interface {
	// Export 导出评审结果，返回导出文件路径和错误
	Export(history *review.ReviewHistory) (string, error)
}

// Format 导出格式
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatHTML     Format = "html"
	FormatPDF      Format = "pdf"
)
