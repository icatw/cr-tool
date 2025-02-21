package review

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/icatw/cr-tool/pkg/config"
)

// Cache 缓存管理器
type Cache struct {
	config *config.CacheConfig
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Content  string    `json:"content"`
	Result   string    `json:"result"`
	DateTime time.Time `json:"datetime"`
}

// NewCache 创建新的缓存管理器
func NewCache() *Cache {
	return &Cache{
		config: &config.Get().Cache,
	}
}

// Get 获取缓存内容
func (c *Cache) Get(content string) string {
	if !c.config.Enabled {
		return ""
	}

	hash := calculateHash(content)
	cacheFile := filepath.Join(c.config.Dir, hash+".json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return ""
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return ""
	}

	// 检查是否过期
	if time.Since(entry.DateTime).Hours() > float64(c.config.ExpireDays*24) {
		os.Remove(cacheFile)
		return ""
	}

	// 验证内容是否匹配
	if entry.Content != content {
		return ""
	}

	return entry.Result
}

// Set 设置缓存内容
func (c *Cache) Set(content, result string) error {
	if !c.config.Enabled {
		return nil
	}

	if err := os.MkdirAll(c.config.Dir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	entry := CacheEntry{
		Content:  content,
		Result:   result,
		DateTime: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	hash := calculateHash(content)
	cacheFile := filepath.Join(c.config.Dir, hash+".json")
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return nil
}

// Clean 清理过期缓存
func (c *Cache) Clean() error {
	if !c.config.Enabled {
		return nil
	}

	entries, err := os.ReadDir(c.config.Dir)
	if err != nil {
		return fmt.Errorf("读取缓存目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(c.config.Dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var cacheEntry CacheEntry
		if err := json.Unmarshal(data, &cacheEntry); err != nil {
			os.Remove(filePath)
			continue
		}

		if time.Since(cacheEntry.DateTime).Hours() > float64(c.config.ExpireDays*24) {
			os.Remove(filePath)
		}
	}

	return nil
}
