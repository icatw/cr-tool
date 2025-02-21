#!/bin/bash
set -euo pipefail

# 加载配置
CONFIG_FILE="./conf/config.json"
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "错误: 配置文件不存在"
    exit 1
fi

# 获取输出目录配置
OUTPUT_DIR=$(jq -r '.output.dir // "./review_results"' "$CONFIG_FILE")

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

# 生成输出文件路径
OUTPUT_FILE="$OUTPUT_DIR/$(date +'%Y%m%d_%H%M%S')_review.md"

# 调用审查工具
REVIEW=$(echo "$DIFF_CONTENT" | go run ./tools/qianwen_review.go)

# 保存审查结果
echo "Code Review Results:" > "$OUTPUT_FILE"
echo "$REVIEW" >> "$OUTPUT_FILE"
echo "Review results saved to $OUTPUT_FILE"

exit 0
