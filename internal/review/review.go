package review

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/your/project/internal/config"
	"github.com/your/project/internal/model"
)

type Reviewer struct {
	config *config.Config
	cache  *Cache
}

func New() *Reviewer {
	return &Reviewer{
		config: config.Get(),
		cache:  NewCache(),
	}
}

func (r *Reviewer) Review(diffContent string) (*model.ReviewHistory, error) {
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
	r.cache.Set(diffContent, result)

	return r.createHistory(diffContent, result)
}

// ... 其他评审相关方法 ...
