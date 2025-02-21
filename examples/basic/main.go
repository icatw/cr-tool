package main

import (
	"fmt"
	"log"

	"github.com/icatw/cr-tool/pkg/config"
	"github.com/icatw/cr-tool/pkg/review"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 创建评审器
	reviewer := review.New()

	// 评审代码
	diffContent := `diff --git a/main.go b/main.go
+ func main() {
+    // TODO: 实现主函数
+ }
`
	history, err := reviewer.Review(diffContent)
	if err != nil {
		log.Fatalf("评审失败: %v", err)
	}

	// 打印评审结果
	fmt.Printf("评审结果：\n%s\n", history.ReviewResult)
}
