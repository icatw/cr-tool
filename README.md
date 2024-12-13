# Git代码评审与通知工具

本项目提供了一种自动化代码评审工具，用于对Git `diff` 的代码改动进行审查，并将结果发送到钉钉Webhook进行通知。它包括一个Bash脚本用于集成到Git工作流程，以及一个Go程序用于通过API执行代码评审并发送通知。

## 功能特性

- 自动获取未提交的`git diff`代码改动。
- 使用配置的AI API（通义千问）执行代码评审。
- 将代码评审结果通过钉钉Webhook发送通知。
- 将评审结果本地保存以供进一步检查。

---

## 快速开始

### 前置要求

- 已安装Git。
- 已安装Go语言开发环境。
- 钉钉Webhook URL及其密钥。
- 通义千问或类似服务的API密钥。

---

### 目录结构

```plaintext
.
├── conf/
│   ├── config.json           # 包含API密钥和Webhook设置的配置文件。
├── review_results/           # 用于保存评审结果的目录。
├── tools/
│   ├── qianwen_review.go     # 用于执行代码评审的Go程序。
├── review.sh                 # 触发评审流程的Bash脚本。
└── README.md                 # 项目文档。
```

---

### 安装

1. 克隆此仓库：

   ```bash
   git clone <repository_url>
   cd <repository_name>
   ```

2. 进入`conf`目录并根据示例创建`config.json`文件：

   ```bash
   cp conf/config.example.json conf/config.json
   ```

3. 使用您的API密钥、钉钉Webhook URL和密钥更新`conf/config.json`。

---

### 配置说明

`config.json`文件的结构如下：

```json
{
    "api_key": "your_api_key_here",
    "model_name": "qwen-plus",
    "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
    "ding_webhook": "https://oapi.dingtalk.com/robot/send?access_token=your_access_token_here",
    "ding_secret": "your_ding_secret_here"
}
```

- **`api_key`**：用于AI评审服务的API密钥。
- **`model_name`**：使用的模型名称，例如`qwen-plus`。
- **`base_url`**：AI评审服务的API地址。
- **`ding_webhook`**：钉钉Webhook URL，用于发送通知。
- **`ding_secret`**：钉钉Webhook的密钥，用于认证。

---

### 使用方法

1. 确保Go程序已编译并可用：

   ```bash
   cd tools
   go build -o qianwen_review qianwen_review.go
   ```

2. 赋予Shell脚本可执行权限：

   ```bash
   chmod +x review.sh
   ```

3. 运行脚本触发评审流程：

   ```bash
   ./review.sh
   ```

4. 结果将保存到`review_results`目录，如果配置正确，也会发送到钉钉。

---

### 示例工作流程

1. 修改Git仓库中的一些文件。
2. 使用`git add`暂存改动。
3. 运行`./review.sh`对暂存改动进行评审。
4. 在`review_results/`中检查评审结果，或通过钉钉查看。

---

## 项目组件

### 1. Shell脚本（`review.sh`）

- 使用`git diff`提取未提交的改动。
- 调用Go程序执行代码评审。
- 将结果保存到`review_results`。

### 2. Go程序（`qianwen_review.go`）

- 将代码差异发送到AI API进行评审。
- 处理并格式化API响应。
- 将格式化后的结果发送到钉钉。

---

## 贡献指南

欢迎贡献！请fork此仓库并提交您的更改。

---

## 许可证

本项目使用MIT许可证，详情请参阅LICENSE文件。

---

## 致谢

- [钉钉开发者文档](https://open.dingtalk.com/document/orgapp/custom-robot-access)
- [通义千问API文档](https://www.aliyun.com/product/ai)
