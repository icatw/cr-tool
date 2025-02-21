package exporter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/icatw/cr-tool/pkg/config"
	"github.com/icatw/cr-tool/pkg/review"
)

type PDFExporter struct {
	config       *config.Config
	htmlExporter *HTMLExporter
}

func NewPDFExporter() *PDFExporter {
	return &PDFExporter{
		config:       config.Get(),
		htmlExporter: NewHTMLExporter(),
	}
}

func (e *PDFExporter) Export(history *review.ReviewHistory) (string, error) {
	// 首先生成 HTML
	htmlPath, err := e.htmlExporter.Export(history)
	if err != nil {
		return "", fmt.Errorf("生成 HTML 失败: %w", err)
	}
	defer os.Remove(htmlPath) // 清理临时 HTML 文件

	// 创建输出目录
	outputDir := e.config.Output.Dir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 设置输出文件路径
	filename := fmt.Sprintf("%s_review.pdf", time.Now().Format("20060102_150405"))
	outputPath := filepath.Join(outputDir, filename)

	// 创建 Chrome 实例
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 生成 PDF
	var pdfData []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf("file://%s", htmlPath)),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(true).
				WithMarginTop(0.4).
				WithMarginBottom(0.4).
				WithMarginLeft(0.4).
				WithMarginRight(0.4).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfData = buf
			return nil
		}),
	); err != nil {
		return "", fmt.Errorf("生成 PDF 失败: %w", err)
	}

	// 保存 PDF 文件
	if err := os.WriteFile(outputPath, pdfData, 0644); err != nil {
		return "", fmt.Errorf("保存 PDF 文件失败: %w", err)
	}

	return outputPath, nil
}
