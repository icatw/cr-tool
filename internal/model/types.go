package model

import "time"

// ReviewStats 评审统计信息
type ReviewStats struct {
	FilesChanged   int            `json:"files_changed"`
	LinesAdded     int            `json:"lines_added"`
	LinesDeleted   int            `json:"lines_deleted"`
	IssuesByLevel  map[string]int `json:"issues_by_level"`
	CommonIssues   []string       `json:"common_issues"`
	ReviewDateTime time.Time      `json:"review_datetime"`
}

// GitInfo Git 信息
type GitInfo struct {
	Branch        string   `json:"branch"`
	CommitHash    string   `json:"commit_hash"`
	CommitMessage string   `json:"commit_message"`
	Author        string   `json:"author"`
	ChangedFiles  []string `json:"changed_files"`
}

// ReviewHistory 评审历史记录
type ReviewHistory struct {
	ID           string       `json:"id"`
	GitInfo      *GitInfo     `json:"git_info"`
	ReviewStats  *ReviewStats `json:"stats"`
	ReviewResult string       `json:"result"`
	DateTime     time.Time    `json:"datetime"`
}

// Message AI 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// RequestBody AI 请求体
type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ResponseBody AI 响应体
type ResponseBody struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
