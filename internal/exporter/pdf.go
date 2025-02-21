package exporter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cr/internal/config"
	"github.com/cr/internal/model"
)

type PDFExporter struct {
	htmlExporter *HTMLExporter
	pageSize     string
	lineNumbers  bool
	highlight    bool
}

func NewPDFExporter() *PDFExporter {
	cfg := config.Get().Output.Reports.PDFOptions
	return &PDFExporter{
		htmlExporter: NewHTMLExporter(),
		pageSize:     cfg.PageSize,
		lineNumbers:  cfg.WithLineNumbers,
		highlight:    cfg.HighlightChanges,
	}
}

func (e *PDFExporter) Export(history *model.ReviewHistory) (string, error) {
	// 首先生成 HTML
	html, err := e.htmlExporter.Export(history)
	if err != nil {
		return "", fmt.Errorf("生成 HTML 失败: %w", err)
	}

	// 创建临时文件
	tmpDir := os.TempDir()
	htmlFile := filepath.Join(tmpDir, fmt.Sprintf("review_%s.html", history.ID))
	pdfFile := filepath.Join(tmpDir, fmt.Sprintf("review_%s.pdf", history.ID))

	// 保存 HTML 文件
	if err := os.WriteFile(htmlFile, []byte(html), 0644); err != nil {
		return "", fmt.Errorf("保存临时 HTML 文件失败: %w", err)
	}
	defer os.Remove(htmlFile)

	// 构建 wkhtmltopdf 命令
	args := []string{
		"--page-size", e.pageSize,
		"--margin-top", "20",
		"--margin-right", "20",
		"--margin-bottom", "20",
		"--margin-left", "20",
		"--encoding", "UTF-8",
	}

	if e.lineNumbers {
		args = append(args, "--footer-right", "[page]/[topage]")
	}

	args = append(args, htmlFile, pdfFile)

	// 执行转换
	cmd := exec.Command("wkhtmltopdf", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("PDF 转换失败: %s, %w", output, err)
	}

	return pdfFile, nil
}
