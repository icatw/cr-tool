#!/bin/bash
set -euo pipefail  # 添加错误处理

# 添加配置文件检查
CONFIG_FILE="./conf/config.json"
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "错误: 配置文件不存在，请复制 config.example.json 到 config.json 并完成配置"
    exit 1
fi

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

# 添加执行结果检查
if ! REVIEW=$(echo "$DIFF_CONTENT" | go run ./tools/qianwen_review.go); then
    echo "代码评审执行失败"
    exit 1
fi

# 保存审查结果
echo "Code Review Results:" > "$OUTPUT_FILE"
echo "$REVIEW" >> "$OUTPUT_FILE"
echo "Review results saved to $OUTPUT_FILE"

exit 0
