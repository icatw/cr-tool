package exporter

import (
	"fmt"
	"github.com/your/project/internal/model"
)

type Exporter interface {
	Export(history *model.ReviewHistory) (string, error)
}

func New(format string) (Exporter, error) {
	switch format {
	case "markdown":
		return NewMarkdownExporter(), nil
	case "html":
		return NewHTMLExporter(), nil
	case "pdf":
		return NewPDFExporter(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
