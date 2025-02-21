package cmd

import (
	"fmt"
	"os"

	"github.com/icatw/cr-tool/pkg/config"
	
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件",
	Long: `初始化配置文件，包括：
- API 密钥配置
- 输出格式配置
- 缓存配置等`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取 API Key
		fmt.Print("请输入您的 API Key: ")
		var apiKey string
		fmt.Scanln(&apiKey)

		if apiKey == "" {
			return fmt.Errorf("API Key 不能为空")
		}

		// 初始化配置
		if err := config.InitConfig(apiKey); err != nil {
			return fmt.Errorf("初始化配置失败: %w", err)
		}

		fmt.Println("配置文件已创建：~/.cr-tool/config.json")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
