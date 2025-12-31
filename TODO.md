# CF-Tool 浏览器桥接移植计划

> **目标**: 将 CF-Tool 的 HTTP 请求迁移到通过用户浏览器发送，以绕过 Cloudflare 保护
>
> **方案**: 使用 MCP-Chrome 项目作为浏览器自动化后端
>
> **日期**: 2025-12-31

---

## 📊 项目概述

### 问题背景
CF-Tool 使用 Go 标准库 `net/http` 直接发送 HTTP 请求到 Codeforces。由于 Cloudflare 的保护机制（JavaScript 挑战、浏览器指纹检测等），直接请求已被阻止。

### 解决方案
利用用户已登录的 Chrome 浏览器，通过 MCP-Chrome 扩展发送请求，从而：
- ✅ 绕过 Cloudflare 保护
- ✅ 复用用户现有的登录会话
- ✅ 无需处理复杂的反爬虫机制

---

## 🏗️ 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                     CF-Tool (Go CLI)                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐      ┌─────────────────────────────────┐ │
│  │ 高层逻辑      │      │   浏览器桥接层 (新增)             │ │
│  │              │      │                                  │ │
│  │ - login      │─────▶│ - BrowserClient                 │ │
│  │ - submit     │      │   - MCP 协议客户端               │ │
│  │ - parse      │      │   - 请求转换器                   │ │
│  │ - watch      │      │   - 响应解析器                   │ │
│  └──────────────┘      └─────────────┬───────────────────┘ │
│                                      │                      │
│                                      │ stdio/HTTP           │
└──────────────────────────────────────┼──────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────┐
│                 MCP Chrome Server                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐      ┌─────────────────────────────────┐ │
│  │ MCP Server   │      │   Chrome Extension               │ │
│  │              │      │                                  │ │
│  │ - 工具注册    │◀─────│ - chrome_network_request        │ │
│  │ - 消息路由    │      │ - chrome_get_web_content        │ │
│  └──────────────┘      │ - chrome_navigate                │ │
│                        └─────────────┬───────────────────┘ │
│                                      │                      │
└──────────────────────────────────────┼──────────────────────┘
                                       │
                                       │ Native Messaging
                                       ▼
                              ┌─────────────────┐
                              │   Chrome 浏览器  │
                              │   (已登录状态)   │
                              └─────────────────┘
```

---

## 📋 实施计划

### 🔨 开发工作流程

**每个功能开发完成后必须执行以下步骤**:

1. **编译测试**:
   ```bash
   make build
   # 或
   go build -o bin/cf ./cmd/cf.go
   ```

2. **功能测试**:
   - 测试新开发的功能是否正常工作
   - 确保不影响现有功能
   - 浏览器模式/HTTP模式切换测试

3. **提交代码**:
   ```bash
   git add .
   git commit -m "feat: [阶段X] 功能描述

   - 实现细节1
   - 实现细节2
   - 测试通过"
   ```

4. **更新 TODO.md**:
   - 在对应任务的 `[ ]` 改为 `[x]`
   - 在"开发日志"中记录进度

**提交信息规范**:
- `feat:` - 新功能
- `fix:` - Bug修复
- `refactor:` - 代码重构
- `docs:` - 文档更新
- `test:` - 测试相关
- `chore:` - 构建/工具相关

**示例**:
```bash
git commit -m "feat: [阶段1] 实现Go MCP客户端基础库

- 实现 pkg/mcp/client.go 核心接口
- 实现 stdio 传输协议
- 实现 JSON-RPC 2.0 消息格式
- 添加错误处理和超时机制
- 编译通过 ✓
- 基础测试通过 ✓"
```

---

### 阶段 1：基础设施搭建 ⏳

#### 1.1 Go MCP 客户端库

**目录结构**:
```
pkg/mcp/
├── client.go          # MCP 协议客户端
├── transport.go       # stdio/HTTP 传输层
├── tools.go           # 工具调用封装
└── types.go           # 数据类型定义
```

**核心接口设计**:
```go
// pkg/mcp/client.go
type Client interface {
    // 调用 MCP 工具
    CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error)

    // 关闭连接
    Close() error

    // 检查连接状态
    Ping(ctx context.Context) error
}

type ToolResult struct {
    Content []interface{} `json:"content"`
    IsError bool          `json:"isError"`
    Data    []byte        `json:"data"`    // 原始响应数据
}
```

**实现要点**:
- [ ] 实现 stdio 传输协议（与 MCP 服务器通信）
- [ ] 实现 JSON-RPC 2.0 消息格式
- [ ] 支持超时和重试机制
- [ ] 错误处理和日志记录

#### 1.2 浏览器 HTTP 客户端

**目录结构**:
```
client/browser/
├── browser_client.go  # 浏览器客户端实现
├── request.go         # 请求封装
├── response.go        # 响应解析
└── cookies.go         # Cookie 管理（可选）
```

**核心接口设计**:
```go
// client/browser/browser_client.go
type BrowserClient struct {
    mcpClient mcp.Client
    timeout   time.Duration
}

// 兼容现有 http.Client 接口
type HttpClient interface {
    Get(url string) (*http.Response, error)
    PostForm(url string, data url.Values) (*http.Response, error)
    Do(req *http.Request) (*http.Response, error)
}

// 新增方法
func NewBrowserClient(mcpClient mcp.Client) *BrowserClient
func (c *BrowserClient) Get(url string) ([]byte, error)
func (c *BrowserClient) Post(url string, data url.Values) ([]byte, error)
func (c *BrowserClient) GetJSON(url string) (map[string]interface{}, error)
```

**实现要点**:
- [ ] 通过 `chrome_network_request` 工具发送请求
- [ ] 将浏览器响应转换为 Go HTTP Response
- [ ] 处理重定向、cookies 等细节
- [ ] 保持与现有 `util.GetBody/PostBody` 兼容

---

### 阶段 2：MCP-Ping 测试工具 🔧

#### 2.1 新增命令: `cf mcp-ping`

**功能**:
- 检测 MCP Chrome Server 是否正确安装
- 测试与浏览器的连接状态
- 显示可用的 MCP 工具列表
- 提供安装提示（如果未安装）

**目录结构**:
```
cmd/
└── mcp-ping.go        # 新增命令
```

**实现**:
```go
// cmd/mcp-ping.go
package cmd

import (
    "fmt"
    "context"
    "time"

    "github.com/NetWilliam/cf-tool/pkg/mcp"
    "github.com/fatih/color"
)

var cmdMcpPing = &Command{
    Usage: "mcp-ping",
    Short: "Test MCP Chrome server connection",
    Long: `
Test if the MCP Chrome server is properly installed and accessible.
This command will:
  1. Try to connect to the MCP server
  2. List available browser tools
  3. Report connection status
  4. Provide installation hints if needed
`,
    Run: mcpPing,
}

func mcpPing(args []string) error {
    color.Cyan("Testing MCP Chrome server connection...\n")

    // 尝试创建 MCP 客户端
    client, err := mcp.NewClient(mcp.Config{
        Transport: "stdio",
        Command:   "node",
        Args:      []string{"/path/to/mcp-chrome-bridge/dist/mcp/mcp-server-stdio.js"},
        Timeout:   5 * time.Second,
    })

    if err != nil {
        color.Red("❌ Failed to create MCP client: %v", err)
        printInstallationHints()
        return err
    }
    defer client.Close()

    // Ping 测试
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := client.Ping(ctx); err != nil {
        color.Red("❌ MCP server ping failed: %v", err)
        printInstallationHints()
        return err
    }

    color.Green("✓ MCP server is running!\n")

    // 获取可用工具列表
    tools, err := client.ListTools(ctx)
    if err != nil {
        color.Yellow("⚠ Could not list tools: %v", err)
        return nil
    }

    color.Cyan("Available tools:")
    for _, tool := range tools {
        color.White("  • %s: %s", tool.Name, tool.Description)
    }

    color.Green("\n✓ Your browser is ready to use with CF-Tool!")
    return nil
}

func printInstallationHints() {
    color.Cyan("\n📦 Installation Guide:")
    color.White(`
1. Install Chrome Extension:
   - Download from: https://github.com/hangwin/mcp-chrome/releases
   - Load in Chrome: chrome://extensions/ → Developer mode → Load unpacked

2. Install Native Host:
   - Follow: https://github.com/hangwin/mcp-chrome/blob/master/docs/INSTALL.md

3. Verify Installation:
   - Run: cf mcp-ping

For more details, visit: https://github.com/hangwin/mcp-chrome
`)
}
```

**集成到主命令**:
```go
// cmd/cf.go
var commands = []*Command{
    // ... 现有命令
    cmdMcpPing,  // 新增
}
```

**测试要点**:
- [ ] 未安装 MCP 时给出清晰提示
- [ ] 已安装时显示可用工具
- [ ] 超时处理（5-10秒）
- [ ] 跨平台兼容性（Windows/Linux/macOS）

---

### 阶段 3：核心功能迁移 🔄

#### 3.1 登录模块重构

**新增文件**: `client/login_browser.go`

**核心流程**:
```
1. chrome_navigate("https://codeforces.com/enter")
2. chrome_get_web_content() → 检测登录状态
3. 如果未登录:
   a. chrome_fill_or_select("#handleOrEmail", username)
   b. chrome_fill_or_select("#password", password)
   c. chrome_click_element("input[type='submit']")
4. 验证登录成功
```

**实现**:
```go
// client/login_browser.go
func (c *Client) LoginWithBrowser() error {
    // 导航到登录页
    if err := c.browser.Navigate(c.host + "/enter"); err != nil {
        return err
    }

    // 检查是否已登录
    logged, handle := c.checkLoginStatus()
    if logged {
        color.Green("Already logged in as %s", handle)
        c.Handle = handle
        return nil
    }

    // 提示用户手动登录（推荐）
    color.Cyan("Please login in the browser within 60 seconds...")
    if err := c.waitForLogin(); err != nil {
        return fmt.Errorf("login timeout: %w", err)
    }

    return nil
}

func (c *Client) checkLoginStatus() (bool, string) {
    // 使用 chrome_get_web_content 获取页面
    // 解析查找用户名
    // 如果找到返回 (true, username)
    // 否则返回 (false, "")
}

func (c *Client) waitForLogin() error {
    // 轮询检查登录状态，最多等待 60 秒
    for i := 0; i < 60; i++ {
        time.Sleep(time.Second)
        if logged, handle := c.checkLoginStatus(); logged {
            c.Handle = handle
            return nil
        }
    }
    fmt.Errorf("login timeout")
}
```

**迁移要点**:
- [ ] 保留现有的加密密码逻辑（可选自动登录）
- [ ] 支持手动登录模式（用户在浏览器中操作）
- [ ] 登录状态持久化到 session 文件

#### 3.2 提交模块重构

**新增文件**: `client/submit_browser.go`

**核心流程**:
```
1. chrome_navigate(submitURL)
2. chrome_get_web_content() → 提取 CSRF token
3. 提取表单字段:
   - ftaa, bfaa
   - programTypeId (语言)
   - source (代码)
4. 填写表单:
   - chrome_fill_or_select("#programTypeId", langID)
   - 注入代码（通过 JavaScript）
5. chrome_click_element("input[type='submit']")
6. 提交后跳转到 mysubmissions
7. WatchSubmission() 监控结果
```

**实现**:
```go
// client/submit_browser.go
func (c *Client) SubmitWithBrowser(info Info, langID, source string) error {
    url := fmt.Sprintf("%s/contest/%d/submit", c.host, info.ContestID)

    // 导航到提交页面
    if err := c.browser.Navigate(url); err != nil {
        return err
    }

    // 获取页面并提取 CSRF
    content, err := c.browser.GetContent()
    if err != nil {
        return err
    }

    csrf := findCsrf(content)

    // 使用 JavaScript 注入代码并提交
    js := fmt.Sprintf(`
        (function() {
            document.querySelector('[name="programTypeId"]').value = "%s";
            document.querySelector('[name="source"]').value = %s;
            document.querySelector('input[type="submit"]').click();
        })();
    `, langID, jsString(source))

    if err := c.browser.ExecuteJS(js); err != nil {
        return err
    }

    // 监控提交状态（复用现有逻辑）
    return c.WatchSubmission(info, 1, true)
}

// jsString 将 Go 字符串转义为 JavaScript 字符串字面量
func jsString(s string) string {
    b, _ := json.Marshal(s)
    return string(b)
}
```

**迁移要点**:
- [ ] CSRF token 提取逻辑复用
- [ ] 代码注入使用 JavaScript 更可靠
- [ ] 保持与原有 `WatchSubmission` 兼容
- [ ] 错误处理和重试机制

#### 3.3 解析模块重构

**修改文件**: `client/parse.go`

**变更**:
```go
// 原代码
body, err := util.GetBody(c.client, URL)

// 新代码
var body []byte
if c.browser != nil {
    body, err = c.browser.GetContentURL(URL)
} else {
    body, err = util.GetBody(c.client, URL)
}
```

**迁移要点**:
- [ ] 最小改动，保持解析逻辑不变
- [ ] 只是替换数据获取方式
- [ ] HTML 解析和样例提取逻辑复用

#### 3.4 其他模块重构

**涉及文件**:
- `client/watch.go` - 监控提交状态
- `client/pull.go` - 拉取代码
- `client/clone.go` - 克隆用户提交
- `client/statis.go` - 获取比赛统计
- `client/race.go` - 比赛倒计时

**重构策略**:
```go
// 通用模式：条件判断使用浏览器还是HTTP
var body []byte
if c.browser != nil {
    // 使用浏览器客户端
    body, err = c.browser.Get(url)
} else {
    // 使用传统HTTP客户端
    body, err = util.GetBody(c.client, url)
}
```

**具体修改**:

1. **client/watch.go**:
```go
// getSubmissions() 函数
// 原代码: util.GetBody(c.client, URL)
// 新代码: 根据c.browser判断使用哪个客户端
```

2. **client/pull.go**:
```go
// PullCode() 函数
// 原代码: util.GetBody(c.client, URL)
// 新代码: 根据c.browser判断使用哪个客户端
```

3. **client/clone.go**:
```go
// Clone() 函数
// 原代码: util.GetJSONBody(c.client, url)
// 新代码: 使用对应的GetJSON方法
```

4. **client/statis.go**:
```go
// Statis() 函数
// 原代码: util.GetBody(c.client, url)
// 新代码: 根据c.browser判断使用哪个客户端
```

5. **client/race.go**:
```go
// RaceContest() 函数
// 原代码: util.GetBody(c.client, url)
// 新代码: 根据c.browser判断使用哪个客户端
```

**迁移要点**:
- [ ] 保持接口不变，只替换底层HTTP调用
- [ ] 统一错误处理
- [ ] 保持向后兼容（HTTP模式仍可用）
- [ ] 浏览器模式下优先使用chrome_get_web_content获取页面

---

### 阶段 4：配置系统 ⚙️

#### 4.1 配置文件更新

**文件**: `~/.cf/config` (JSON 格式)

**新增字段**:
```json
{
  "aliases": [],
  "testcases": [],
  "default": {
    "language": "54"
  },
  "browser": {
    "enabled": true,
    "mcp_transport": "stdio",
    "mcp_command": "node",
    "mcp_args": [
      "/path/to/mcp-chrome-bridge/dist/mcp/mcp-server-stdio.js"
    ],
    "auto_login": false,
    "fallback_to_http": false
  }
}
```

**配置结构**:
```go
// config/config.go
type Config struct {
    // ... 现有字段

    Browser BrowserConfig `json:"browser"`
}

type BrowserConfig struct {
    // 是否启用浏览器模式
    Enabled bool `json:"enabled"`

    // MCP 传输方式: "stdio" 或 "http"
    Transport string `json:"mcp_transport"`

    // MCP 服务器命令
    Command string `json:"mcp_command"`

    // MCP 服务器参数
    Args []string `json:"mcp_args"`

    // 是否自动登录（false = 手动登录）
    AutoLogin bool `json:"auto_login"`

    // 浏览器失败时是否回退到 HTTP
    FallbackToHTTP bool `json:"fallback_to_http"`
}
```

#### 4.2 配置命令更新

**修改**: `cmd/config.go`

**新增交互**:
```go
func configBrowser() {
    color.Cyan("\n--- Browser Configuration ---")

    enabled := util.YesOrNo("Enable browser mode (recommended for Cloudflare)?")
    if !enabled {
        return
    }

    color.Cyan("\n✓ Browser mode enabled!")
    color.White("Please ensure MCP Chrome Server is installed:")
    color.White("  Run: cf mcp-ping to test\n")

    // 其他配置项...
}
```

---

### 阶段 5：部署和文档 📚

#### 5.1 安装脚本

**文件**: `scripts/install-browser.sh`

```bash
#!/bin/bash
set -e

echo "🔧 CF-Tool Browser Mode Setup"
echo "================================"

# 1. 检测 Chrome
if command -v google-chrome &> /dev/null; then
    CHROME=$(command -v google-chrome)
elif command -v chromium &> /dev/null; then
    CHROME=$(command -v chromium)
else
    echo "❌ Chrome not found. Please install Chrome first."
    exit 1
fi

echo "✓ Found Chrome: $CHROME"

# 2. 下载 MCP-Chrome
MCP_DIR="$HOME/.mcp-chrome"
mkdir -p "$MCP_DIR"

echo "📦 Downloading MCP Chrome Server..."
cd "$MCP_DIR"
git clone https://github.com/hangwin/mcp-chrome.git
cd mcp-chrome
pnpm install
pnpm build

# 3. 安装扩展
echo "🔌 Installing Chrome Extension..."
EXT_PATH="$MCP_DIR/mcp-chrome/app/chrome-extension"
echo "Please load the following directory in Chrome:"
echo "  chrome://extensions/ → Developer mode → Load unpacked"
echo "  Path: $EXT_PATH"
read -p "Press Enter after installing the extension..."

# 4. 安装 Native Host
echo "🔧 Installing Native Host..."
cd "$MCP_DIR/mcp-chrome/app/native-server"
pnpm run register

# 5. 测试连接
echo "🧪 Testing connection..."
cf mcp-ping

echo ""
echo "✅ Installation complete!"
echo "You can now use: cf submit"
```

#### 5.2 文档更新

**修改**: `README.md`

**新增章节**:
```markdown
## Browser Mode (Recommended)

### Why Browser Mode?

Codeforces uses Cloudflare protection which blocks direct HTTP requests.
Browser mode uses your installed Chrome browser to bypass these restrictions.

### Installation

1. Install MCP Chrome Server:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/NetWilliam/cf-tool/master/scripts/install-browser.sh | bash
   ```

2. Test installation:
   ```bash
   cf mcp-ping
   ```

3. Enable browser mode:
   ```bash
   cf config
   # Choose "Enable browser mode: yes"
   ```

### Usage

```bash
# Submit code (browser will handle Cloudflare)
cf submit 1234 A main.cpp

# Parse problem
cf parse 1234 A

# Pull submissions
cf pull 1234
```
```

---

## 🔄 迁移对照表

| 原 CF-Tool 函数 | 新实现 | MCP 工具映射 |
|----------------|--------|-------------|
| `util.GetBody(client, url)` | `browserClient.Get(url)` | `chrome_network_request` (GET) |
| `util.PostBody(client, url, data)` | `browserClient.Post(url, data)` | `chrome_network_request` (POST) |
| `util.GetJSONBody(client, url)` | `browserClient.GetJSON(url)` | `chrome_network_request` + JSON 解析 |
| `c.Login()` | `c.LoginWithBrowser()` | `chrome_navigate` + `chrome_get_web_content` |
| `c.Submit()` | `c.SubmitWithBrowser()` | `chrome_navigate` + JS 注入 |
| Cookie Jar 管理 | 浏览器自动管理 | 无需额外处理 |
| HTML 解析逻辑 | 复用现有代码 | `chrome_get_web_content` |

---

## 📦 交付物清单

- [ ] **阶段 1**: Go MCP 客户端库 (`pkg/mcp/`)
- [ ] **阶段 1**: 浏览器 HTTP 客户端 (`client/browser/`)
- [ ] **阶段 2**: `cf mcp-ping` 测试命令 (`cmd/mcp-ping.go`)
- [ ] **阶段 3**: 重构登录模块 (`client/login_browser.go`)
- [ ] **阶段 3**: 重构提交模块 (`client/submit_browser.go`)
- [ ] **阶段 3**: 重构解析模块（修改 `client/parse.go`）
- [ ] **阶段 3**: 重构其他模块（watch, pull, clone）
- [ ] **阶段 4**: 配置文件格式更新 (`config/config.go`)
- [ ] **阶段 4**: 配置命令更新 (`cmd/config.go`)
- [ ] **阶段 5**: 安装脚本 (`scripts/install-browser.sh`)
- [ ] **阶段 5**: 用户文档更新 (`README.md`)
- [ ] **阶段 5**: 升级脚本（自动迁移现有用户）

---

## 🧪 测试计划

### 单元测试
- [ ] MCP 客户端通信测试
- [ ] 请求/响应转换测试
- [ ] Cookie 管理测试

### 集成测试
- [ ] `cf mcp-ping` 测试
- [ ] 完整登录流程测试
- [ ] 提交代码流程测试
- [ ] 解析题目流程测试

### 兼容性测试
- [ ] 保持向后兼容（HTTP 模式仍可用）
- [ ] 跨平台测试（Windows/Linux/macOS）
- [ ] 不同 Chrome 版本测试

---

## ⚠️ 风险和注意事项

### 技术风险
1. **MCP 服务器依赖**: 用户需要安装额外的组件
   - 缓解: 提供一键安装脚本
   - 回退: 保留 HTTP 模式

2. **浏览器版本兼容**: 不同 Chrome 版本可能有差异
   - 缓解: 测试主流版本（Chromium ≥ 90）

3. **性能开销**: 浏览器通信比直接 HTTP 慢
   - 影响: 可接受（提交操作不频繁）

### 用户体验
1. **首次配置复杂**: 需要安装扩展和 Native Host
   - 缓解: 提供详细文档和自动化脚本
   - 测试: `cf mcp-ping` 验证安装

2. **手动登录**: 用户首次需要在浏览器登录
   - 缓解: 登录一次后 session 持久化

---

## 📅 里程碑

- [ ] **M1**: 基础设施完成（MCP 客户端 + 浏览器客户端）
- [ ] **M2**: `cf mcp-ping` 命令可用
- [ ] **M3**: 登录功能迁移完成
- [ ] **M4**: 提交功能迁移完成（核心功能可用）
- [ ] **M5**: 所有功能迁移完成
- [ ] **M6**: 文档和安装脚本完成
- [ ] **M7**: 测试和修复 Bug
- [ ] **M8**: 发布正式版本

---

## 📝 开发日志

### 2025-12-31
- ✅ 完成项目规划
- ✅ 创建详细的 TODO.md
- 🔄 待开始实现

---

## 🔗 参考资源

- [MCP-Chrome GitHub](https://github.com/hangwin/mcp-chrome)
- [MCP Protocol Spec](https://modelcontextprotocol.io/)
- [Chrome Extension Docs](https://developer.chrome.com/docs/extensions/)
- [CF-Tool 原项目](https://github.com/xalanq/cf-tool)

---

**最后更新**: 2025-12-31
**维护者**: @NetWilliam
