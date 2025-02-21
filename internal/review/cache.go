package review

import (
	"encoding/json"
	"github.com/cr/internal/config"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	config *config.CacheConfig
}

type CacheEntry struct {
	Content  string    `json:"content"`
	Result   string    `json:"result"`
	DateTime time.Time `json:"datetime"`
}

func NewCache() *Cache {
	return &Cache{
		config: &config.Get().Cache,
	}
}

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

	return entry.Result
}

func (c *Cache) Set(content, result string) error {
	if !c.config.Enabled {
		return nil
	}

	if err := os.MkdirAll(c.config.Dir, 0755); err != nil {
		return err
	}

	entry := CacheEntry{
		Content:  content,
		Result:   result,
		DateTime: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	hash := calculateHash(content)
	return os.WriteFile(filepath.Join(c.config.Dir, hash+".json"), data, 0644)
}
