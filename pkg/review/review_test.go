package review

import (
	"github.com/icatw/cr-tool/pkg/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReview(t *testing.T) {
	tests := []struct {
		name        string
		diffContent string
		wantErr     error
	}{
		{
			name:        "empty diff",
			diffContent: "",
			wantErr:     ErrEmptyDiff,
		},
		{
			name:        "diff too large",
			diffContent: string(make([]byte, 3000)),
			wantErr:     ErrDiffTooLarge,
		},
		// ... 更多测试用例 ...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			_, err := r.Review(tt.diffContent)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
