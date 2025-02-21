package exporter

import (
	"github.com/icatw/cr-tool/pkg/review"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestMarkdownExporter_Export(t *testing.T) {
	// 创建测试数据
	history := &review.ReviewHistory{
		ID:           "test",
		ReviewResult: "# Test Review\n\nThis is a test review.",
	}

	// 创建导出器
	exp := NewMarkdownExporter()

	// 执行导出
	path, err := exp.Export(history)
	assert.NoError(t, err)
	assert.NotEmpty(t, path)

	// 验证文件内容
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test Review")

	// 清理测试文件
	os.Remove(path)
}
