package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/cr/internal/model"
)

// GetInfo GitInfo 获取 Git 信息
func GetInfo() (*model.GitInfo, error) {
	info := &model.GitInfo{}

	// 获取当前分支
	branch, err := execGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("获取分支信息失败: %w", err)
	}
	info.Branch = branch

	// 获取最近的提交信息
	commit, err := execGitCommand("log", "-1", "--pretty=format:%H|%s|%an")
	if err != nil {
		return nil, fmt.Errorf("获取提交信息失败: %w", err)
	}

	parts := strings.Split(commit, "|")
	if len(parts) == 3 {
		info.CommitHash = parts[0]
		info.CommitMessage = parts[1]
		info.Author = parts[2]
	}

	// 获取变更文件列表
	files, err := getChangedFiles()
	if err != nil {
		return nil, fmt.Errorf("获取变更文件列表失败: %w", err)
	}
	info.ChangedFiles = files

	return info, nil
}

// getChangedFiles 获取变更文件列表
func getChangedFiles() ([]string, error) {
	// 获取暂存区的文件
	staged, err := execGitCommand("diff", "--cached", "--name-only")
	if err != nil {
		return nil, err
	}

	// 获取工作区的文件
	unstaged, err := execGitCommand("diff", "--name-only")
	if err != nil {
		return nil, err
	}

	// 获取未跟踪的文件
	untracked, err := execGitCommand("ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}

	// 合并所有文件列表并去重
	filesMap := make(map[string]bool)
	for _, files := range []string{staged, unstaged, untracked} {
		if files = strings.TrimSpace(files); files != "" {
			for _, file := range strings.Split(files, "\n") {
				filesMap[file] = true
			}
		}
	}

	// 转换为切片
	var result []string
	for file := range filesMap {
		result = append(result, file)
	}

	return result, nil
}

// execGitCommand 执行 Git 命令
func execGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("git command failed: %s", stderr.String())
		}
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

// IsGitRepo 检查当前目录是否是 Git 仓库
func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// GetRepoRoot 获取仓库根目录
func GetRepoRoot() (string, error) {
	return execGitCommand("rev-parse", "--show-toplevel")
}
