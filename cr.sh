#!/bin/bash

# 配置输出目录
OUTPUT_DIR="./review_results"
OUTPUT_FILE="$OUTPUT_DIR/$(date +'%Y%m%d_%H%M%S')_review.md"

# 检查并创建输出目录
if [[ ! -d "$OUTPUT_DIR" ]]; then
    mkdir -p "$OUTPUT_DIR"
fi

# 获取未提交的代码改动
DIFF_CONTENT=$(git diff --unified=0)
if [[ -z "$DIFF_CONTENT" ]]; then
    echo "No changes to review."
    exit 0
fi

# 调用审查工具
REVIEW=$(echo "$DIFF_CONTENT" | go run ./tools/qianwen_review.go)

# 保存审查结果
echo "Code Review Results:" > "$OUTPUT_FILE"
echo "$REVIEW" >> "$OUTPUT_FILE"
echo "Review results saved to $OUTPUT_FILE"

exit 0
