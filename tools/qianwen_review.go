package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 扩展配置结构
type OutputConfig struct {
	Dir    string `json:"dir"`
	Format string `json:"format"`
}

type CacheConfig struct {
	Enabled    bool   `json:"enabled"`
	Dir        string `json:"dir"`
	ExpireDays int    `json:"expire_days"`
}

// 修改配置结构
type DingConfig struct {
	Enabled bool   `json:"enabled"`
	Webhook string `json:"webhook"`
	Secret  string `json:"secret"`
}

// 添加评审模板配置结构
type ReviewTemplate struct {
	SystemPrompt string   `json:"system_prompt"`
	FocusPoints  []string `json:"focus_points"`
}

type ReviewConfig struct {
	Template       string                    `json:"template"`
	Templates      map[string]ReviewTemplate `json:"templates"`
	IgnorePatterns []string                  `json:"ignore_patterns"`
	MaxDiffSize    int                       `json:"max_diff_size"`
}

// Config 配置文件结构体
type Config struct {
	APIKey    string       `json:"api_key"`
	ModelName string       `json:"model_name"`
	BaseURL   string       `json:"base_url"`
	Ding      DingConfig   `json:"ding"`
	Output    OutputConfig `json:"output"`
	Cache     CacheConfig  `json:"cache"`
	Review    ReviewConfig `json:"review"`
}

// 全局配置
var config Config

// 添加缓存结构
type ReviewCache struct {
	Content  string    `json:"content"`
	Result   string    `json:"result"`
	DateTime time.Time `json:"datetime"`
}

// 添加评审统计结构
type ReviewStats struct {
	FilesChanged   int            `json:"files_changed"`
	LinesAdded     int            `json:"lines_added"`
	LinesDeleted   int            `json:"lines_deleted"`
	IssuesByLevel  map[string]int `json:"issues_by_level"`
	CommonIssues   []string       `json:"common_issues"`
	ReviewDateTime time.Time      `json:"review_datetime"`
}

// 添加 Git 相关功能
type GitInfo struct {
	Branch        string
	CommitHash    string
	CommitMessage string
	Author        string
	ChangedFiles  []string
}

// 添加历史记录结构
type ReviewHistory struct {
	ID           string       `json:"id"`
	GitInfo      *GitInfo     `json:"git_info"`
	ReviewStats  *ReviewStats `json:"stats"`
	ReviewResult string       `json:"result"`
	DateTime     time.Time    `json:"datetime"`
}

// 添加 Markdown 格式化结构
type MarkdownReport struct {
	Title       string
	Summary     string
	GitInfo     *GitInfo
	Stats       *ReviewStats
	ReviewItems []string
	DateTime    time.Time
}

// 添加导出格式枚举
const (
	FormatMarkdown = "markdown"
	FormatHTML     = "html"
	FormatPDF      = "pdf"
)

// 添加导出接口
type Exporter interface {
	Export(history *ReviewHistory) (string, error)
}

// Markdown 导出器
type MarkdownExporter struct{}

func (e *MarkdownExporter) Export(history *ReviewHistory) (string, error) {
	return formatMarkdownReport(history), nil
}

// HTML 导出器
type HTMLExporter struct {
	CSSTemplate string
}

func (e *HTMLExporter) Export(history *ReviewHistory) (string, error) {
	var html strings.Builder

	// 添加 HTML 头部
	html.WriteString(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>代码评审报告</title>
	<style>
		/* 可以根据 CSSTemplate 加载不同的样式 */
		body {
			font-family: -apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif;
			line-height: 1.6;
			max-width: 1200px;
			margin: 0 auto;
			padding: 2rem;
		}
		.review-header { margin-bottom: 2rem; }
		.git-info { background: #f6f8fa; padding: 1rem; border-radius: 6px; }
		.stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; }
		.issue { border-left: 4px solid #e36209; padding-left: 1rem; margin: 1rem 0; }
	</style>
</head>
<body>`)

	// 添加标题
	html.WriteString("<h1>代码评审报告</h1>")

	// Git 信息
	if history.GitInfo != nil {
		html.WriteString(`<div class="git-info">
			<h2>Git 信息</h2>
			<p>分支: <code>` + history.GitInfo.Branch + `</code></p>
			<p>提交: <code>` + history.GitInfo.CommitHash + `</code></p>
			<p>作者: ` + history.GitInfo.Author + `</p>
			<p>提交信息: ` + history.GitInfo.CommitMessage + `</p>
		</div>`)
	}

	// ... 其他内容转换为 HTML ...

	html.WriteString("</body></html>")
	return html.String(), nil
}

// PDF 导出器
type PDFExporter struct {
	PageSize         string
	WithLineNumbers  bool
	HighlightChanges bool
}

func (e *PDFExporter) Export(history *ReviewHistory) (string, error) {
	// 首先生成 HTML
	htmlExporter := &HTMLExporter{CSSTemplate: "github"}
	htmlContent, err := htmlExporter.Export(history)
	if err != nil {
		return "", err
	}

	// 使用 wkhtmltopdf 转换为 PDF
	// 这里需要系统安装 wkhtmltopdf
	tmpFile := filepath.Join(os.TempDir(), "review_"+history.ID+".html")
	if err := os.WriteFile(tmpFile, []byte(htmlContent), 0644); err != nil {
		return "", err
	}
	defer os.Remove(tmpFile)

	pdfFile := filepath.Join(os.TempDir(), "review_"+history.ID+".pdf")
	cmd := exec.Command("wkhtmltopdf",
		"--page-size", e.PageSize,
		"--enable-local-file-access",
		tmpFile, pdfFile)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfFile, nil
}

// 导出器工厂
func createExporter(format string) (Exporter, error) {
	switch format {
	case FormatMarkdown:
		return &MarkdownExporter{}, nil
	case FormatHTML:
		return &HTMLExporter{
			CSSTemplate: config.Output.Reports.CSSTemplate,
		}, nil
	case FormatPDF:
		return &PDFExporter{
			PageSize:         config.Output.Reports.PDFOptions.PageSize,
			WithLineNumbers:  config.Output.Reports.PDFOptions.WithLineNumbers,
			HighlightChanges: config.Output.Reports.PDFOptions.HighlightChanges,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// 修改保存报告的函数
func saveReport(history *ReviewHistory) error {
	for _, format := range config.Output.Format {
		exporter, err := createExporter(format)
		if err != nil {
			log.Printf("Failed to create exporter for format %s: %v", format, err)
			continue
		}

		content, err := exporter.Export(history)
		if err != nil {
			log.Printf("Failed to export report in format %s: %v", format, err)
			continue
		}

		// 创建输出目录
		reportDir := filepath.Join(config.Output.Dir, format)
		if err := os.MkdirAll(reportDir, 0755); err != nil {
			return err
		}

		// 保存文件
		filename := filepath.Join(reportDir,
			fmt.Sprintf("%s_%s.%s",
				history.DateTime.Format("20060102_150405"),
				history.ID,
				format))

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			log.Printf("Failed to save report in format %s: %v", format, err)
		} else {
			log.Printf("Saved report in format %s: %s", format, filename)
		}
	}
	return nil
}

// 计算内容的哈希值作为缓存键
func calculateHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 检查缓存
func checkCache(content string) (string, error) {
	if !config.Cache.Enabled {
		return "", nil
	}

	hash := calculateHash(content)
	cacheFile := filepath.Join(config.Cache.Dir, hash+".json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", nil
	}

	var cache ReviewCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return "", nil
	}

	// 检查是否过期
	if time.Since(cache.DateTime).Hours() > float64(config.Cache.ExpireDays*24) {
		os.Remove(cacheFile)
		return "", nil
	}

	// 验证内容是否匹配
	if cache.Content != content {
		return "", nil
	}

	return cache.Result, nil
}

// 保存缓存
func saveCache(content, result string) error {
	if !config.Cache.Enabled {
		return nil
	}

	if err := os.MkdirAll(config.Cache.Dir, 0755); err != nil {
		return err
	}

	cache := ReviewCache{
		Content:  content,
		Result:   result,
		DateTime: time.Now(),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	hash := calculateHash(content)
	return os.WriteFile(filepath.Join(config.Cache.Dir, hash+".json"), data, 0644)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ResponseBody struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// 加载配置文件
func loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	return nil
}

// Generate HMAC-SHA256 签名
func generateSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// 修改发送钉钉消息的函数
func sendDingMessage(message string) error {
	// 如果钉钉通知未启用，直接返回
	if !config.Ding.Enabled {
		log.Println("DingTalk notification is disabled")
		return nil
	}

	// 验证必要的配置
	if config.Ding.Webhook == "" || config.Ding.Secret == "" {
		return fmt.Errorf("DingTalk webhook or secret is not configured")
	}

	timestamp := time.Now().UnixMilli()
	sign, err := generateSign(config.Ding.Secret, timestamp)
	if err != nil {
		return fmt.Errorf("failed to generate DingTalk sign: %w", err)
	}

	// 构造完整的 Webhook URL
	webhookURL := fmt.Sprintf("%s&timestamp=%d&sign=%s",
		config.Ding.Webhook, timestamp, url.QueryEscape(sign))

	// 构造消息内容
	body := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": message,
		},
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal DingTalk message body: %w", err)
	}

	// 发送 HTTP POST 请求
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send DingTalk message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	log.Println("Message sent successfully!")
	return nil
}

// 添加文件过滤功能
func shouldIgnoreFile(filename string) bool {
	for _, pattern := range config.Review.IgnorePatterns {
		matched, err := filepath.Match(pattern, filename)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// 修改代码评审函数
func performCodeReview(diffContent string) (string, error) {
	// 检查 diff 大小
	if len(diffContent) > config.Review.MaxDiffSize {
		return "", fmt.Errorf("diff 内容超过最大限制 (%d > %d bytes)",
			len(diffContent), config.Review.MaxDiffSize)
	}

	// 获取模板
	template, ok := config.Review.Templates[config.Review.Template]
	if !ok {
		template = config.Review.Templates["default"]
	}

	payload := RequestBody{
		Model: config.ModelName,
		Messages: []Message{
			{
				Role:    "system",
				Content: template.SystemPrompt,
			},
			{
				Role:    "user",
				Content: diffContent,
			},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 检查缓存
	if result, err := checkCache(diffContent); err == nil && result != "" {
		log.Println("Using cached review result")
		return result, nil
	}

	// 发送审查请求
	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result ResponseBody
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// 获取结果后保存缓存
	if len(result.Choices) > 0 {
		reviewResult := result.Choices[0].Message.Content
		if err := saveCache(diffContent, reviewResult); err != nil {
			log.Printf("Failed to save cache: %v", err)
		}
		return reviewResult, nil
	}
	return "No review results returned.", nil
}

// 扩展统计分析功能
func analyzeReviewStats(diffContent string, reviewResult string) (*ReviewStats, error) {
	stats := &ReviewStats{
		IssuesByLevel:  make(map[string]int),
		CommonIssues:   make([]string, 0),
		ReviewDateTime: time.Now(),
	}

	// 分析 diff 内容
	var currentFile string
	changedFiles := make(map[string]bool)

	lines := strings.Split(diffContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) > 2 {
				currentFile = strings.TrimPrefix(parts[2], "b/")
				if !shouldIgnoreFile(currentFile) {
					changedFiles[currentFile] = true
				}
			}
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			stats.LinesAdded++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			stats.LinesDeleted++
		}
	}
	stats.FilesChanged = len(changedFiles)

	// 分析评审结果
	sections := strings.Split(reviewResult, "##")
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if strings.HasPrefix(section, "主要问题") {
			// 统计问题级别
			if strings.Contains(section, "严重") {
				stats.IssuesByLevel["严重"]++
			} else if strings.Contains(section, "中等") {
				stats.IssuesByLevel["中等"]++
			} else if strings.Contains(section, "低") {
				stats.IssuesByLevel["低"]++
			}

			// 提取常见问题
			lines := strings.Split(section, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") {
					issue := strings.TrimSpace(strings.TrimPrefix(line, "1."))
					issue = strings.TrimSpace(strings.TrimPrefix(issue, "2."))
					if issue != "" {
						stats.CommonIssues = append(stats.CommonIssues, issue)
					}
				}
			}
		}
	}

	return stats, nil
}

// 获取 Git 信息
func getGitInfo() (*GitInfo, error) {
	gitInfo := &GitInfo{}

	// 获取当前分支
	output, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return nil, err
	}
	gitInfo.Branch = strings.TrimSpace(string(output))

	// 获取最近的提交信息
	output, err = exec.Command("git", "log", "-1", "--pretty=format:%H|%s|%an").Output()
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(output), "|")
	if len(parts) == 3 {
		gitInfo.CommitHash = parts[0]
		gitInfo.CommitMessage = parts[1]
		gitInfo.Author = parts[2]
	}

	return gitInfo, nil
}

// 添加历史记录存储
func saveReviewHistory(history *ReviewHistory) error {
	historyDir := filepath.Join(config.Output.Dir, "history")
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(historyDir,
		fmt.Sprintf("%s_%s.json", history.DateTime.Format("20060102_150405"), history.ID))

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// 格式化评审结果为 Markdown
func formatMarkdownReport(history *ReviewHistory) string {
	var report strings.Builder

	// 添加标题
	report.WriteString(fmt.Sprintf("# 代码评审报告\n\n"))

	// Git 信息
	if history.GitInfo != nil {
		report.WriteString("## Git 信息\n\n")
		report.WriteString(fmt.Sprintf("- 分支: `%s`\n", history.GitInfo.Branch))
		report.WriteString(fmt.Sprintf("- 提交: `%s`\n", history.GitInfo.CommitHash))
		report.WriteString(fmt.Sprintf("- 作者: %s\n", history.GitInfo.Author))
		report.WriteString(fmt.Sprintf("- 提交信息: %s\n", history.GitInfo.CommitMessage))
	}

	// 统计信息
	report.WriteString("\n## 统计信息\n\n")
	report.WriteString(fmt.Sprintf("- 文件变化: %d\n", history.ReviewStats.FilesChanged))
	report.WriteString(fmt.Sprintf("- 新增行数: %d\n", history.ReviewStats.LinesAdded))
	report.WriteString(fmt.Sprintf("- 删除行数: %d\n", history.ReviewStats.LinesDeleted))

	// 问题级别统计
	report.WriteString("\n## 问题级别统计\n\n")
	for level, count := range history.ReviewStats.IssuesByLevel {
		report.WriteString(fmt.Sprintf("- %s: %d\n", level, count))
	}

	// 常见问题
	report.WriteString("\n## 常见问题\n\n")
	for _, issue := range history.ReviewStats.CommonIssues {
		report.WriteString(fmt.Sprintf("- %s\n", issue))
	}

	return report.String()
}

func main() {
	// ... 其他代码保持不变 ...

	// 保存报告（替换原来的 saveMarkdownReport 调用）
	if err := saveReport(history); err != nil {
		log.Printf("Failed to save reports: %v", err)
	}

	// 根据第一个配置的格式显示结果
	if len(config.Output.Format) > 0 {
		exporter, err := createExporter(config.Output.Format[0])
		if err == nil {
			content, err := exporter.Export(history)
			if err == nil {
				fmt.Println(content)
			}
		}
	}

	// ... 其他代码保持不变 ...
}
