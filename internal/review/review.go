package review

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cr/internal/config"
	"github.com/cr/internal/git"
	"github.com/cr/internal/model"
)

// Reviewer 代码评审器
type Reviewer struct {
	config *config.Config
	cache  *Cache
}

// New 创建新的评审器
func New() *Reviewer {
	return &Reviewer{
		config: config.Get(),
		cache:  NewCache(),
	}
}

// Review 执行代码评审
func (r *Reviewer) Review(diffContent string) (*model.ReviewHistory, error) {
	// 检查 diff 大小
	if len(diffContent) > r.config.Review.MaxDiffSize {
		return nil, fmt.Errorf("diff 内容超过最大限制 (%d > %d bytes)",
			len(diffContent), r.config.Review.MaxDiffSize)
	}

	// 检查缓存
	if result := r.cache.Get(diffContent); result != "" {
		return r.createHistory(diffContent, result)
	}

	// 执行评审
	result, err := r.performReview(diffContent)
	if err != nil {
		return nil, err
	}

	// 保存缓存
	if err := r.cache.Set(diffContent, result); err != nil {
		// 仅记录错误，不影响主流程
		fmt.Printf("保存缓存失败: %v\n", err)
	}

	return r.createHistory(diffContent, result)
}

// performReview 执行实际的评审请求
func (r *Reviewer) performReview(diffContent string) (string, error) {
	// 获取模板
	template, ok := r.config.Review.Templates[r.config.Review.Template]
	if !ok {
		template = r.config.Review.Templates["default"]
	}

	payload := model.RequestBody{
		Model: r.config.ModelName,
		Messages: []model.Message{
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
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 发送请求
	req, err := http.NewRequest("POST", r.config.BaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.config.APIKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result model.ResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("未获取到评审结果")
	}

	return result.Choices[0].Message.Content, nil
}

// createHistory 创建评审历史记录
func (r *Reviewer) createHistory(diffContent, result string) (*model.ReviewHistory, error) {
	// 获取 Git 信息
	gitInfo, err := r.getGitInfo()
	if err != nil {
		// 记录错误但继续执行
		fmt.Printf("获取 Git 信息失败: %v\n", err)
	}

	// 分析统计信息
	stats, err := r.analyzeStats(diffContent, result)
	if err != nil {
		fmt.Printf("分析统计信息失败: %v\n", err)
	}

	return &model.ReviewHistory{
		ID:           calculateHash(diffContent)[:8],
		GitInfo:      gitInfo,
		ReviewStats:  stats,
		ReviewResult: result,
		DateTime:     time.Now(),
	}, nil
}

// calculateHash 计算内容的哈希值
func calculateHash(content string) string {
	h := sha256.New()
	h.Write([]byte(content))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// getGitInfo 获取 Git 信息
func (r *Reviewer) getGitInfo() (*model.GitInfo, error) {
	// 检查是否在 Git 仓库中
	if !git.IsGitRepo() {
		return nil, fmt.Errorf("当前目录不是 Git 仓库")
	}

	// 获取 Git 信息
	return git.GetInfo()
}

// ... 其他评审相关方法 ...
