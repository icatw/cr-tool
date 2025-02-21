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

func main() {
	// 加载配置文件
	err := loadConfig("conf/config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 读取标准输入中的代码差异内容
	var diffContent strings.Builder
	_, err = io.Copy(&diffContent, os.Stdin)
	if err != nil {
		log.Fatalf("Failed to read diff content: %v", err)
	}
	// 执行代码评审
	reviewResult, err := performCodeReview(diffContent.String())
	if err != nil {
		log.Printf("Code review failed: %v", err)
		fmt.Print("No valid review result.\n")
		return
	}

	// 获取 Git 信息
	gitInfo, err := getGitInfo()
	if err != nil {
		log.Printf("Failed to get git info: %v", err)
	}

	// 分析统计信息
	stats, err := analyzeReviewStats(diffContent.String(), reviewResult)
	if err != nil {
		log.Printf("Failed to analyze review stats: %v", err)
	}

	// 保存历史记录
	history := &ReviewHistory{
		ID:           calculateHash(diffContent.String())[:8],
		GitInfo:      gitInfo,
		ReviewStats:  stats,
		ReviewResult: reviewResult,
		DateTime:     time.Now(),
	}

	if err := saveReviewHistory(history); err != nil {
		log.Printf("Failed to save review history: %v", err)
	}

	// 输出评审结果
	fmt.Println(reviewResult)

	// 输出统计信息
	fmt.Printf("\n统计信息:\n")
	fmt.Printf("- 变更文件数: %d\n", stats.FilesChanged)
	fmt.Printf("- 新增行数: %d\n", stats.LinesAdded)
	fmt.Printf("- 删除行数: %d\n", stats.LinesDeleted)
	fmt.Printf("- 问题分布:\n")
	for level, count := range stats.IssuesByLevel {
		fmt.Printf("  - %s: %d\n", level, count)
	}

	// 发送钉钉消息（如果启用）
	if err := sendDingMessage(reviewResult); err != nil {
		log.Printf("Failed to send DingTalk message: %v", err)
	}

	log.Println("All tasks completed.")
}
