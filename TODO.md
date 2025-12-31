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

### 阶段 1：基础设施搭建 ✅

#### 1.1 Go MCP 客户端库 ✅

**目录结构**:
```
pkg/mcp/
├── client.go          # MCP 协议客户端
├── tools.go           # 工具调用封装
└── types.go           # 数据类型定义
```

**核心接口**:
- `Client` - MCP 协议客户端
- `CallTool()` - 调用 MCP 工具
- `Close()` - 关闭连接

**已完成**:
- [x] 实现 stdio 传输协议
- [x] 实现 HTTP 传输协议
- [x] 实现 JSON-RPC 2.0 消息格式
- [x] 支持超时和重试机制
- [x] 错误处理和日志记录
- [x] 封装常用 Chrome 工具调用

#### 1.2 浏览器 HTTP 客户端 ✅

**文件**: `client/fetcher.go`

**核心接口**:
```go
type Fetcher interface {
    Get(url string) ([]byte, error)
    GetJSON(url string) (map[string]interface{}, error)
    Post(url string, data url.Values) ([]byte, error)
}

type HTTPFetcher struct { ... }
type BrowserFetcher struct { ... }
```

**已完成**:
- [x] HTTPFetcher - 传统 HTTP 模式
- [x] BrowserFetcher - 浏览器模式
- [x] 通过 `chrome_network_request` 发送请求
- [x] 通过 `chrome_get_web_content` 获取 HTML
- [x] 统一的 Fetcher 接口
- [x] 自动检测并切换模式

---

### 阶段 2：MCP-Ping 测试工具 ✅

#### 2.1 新增命令: `cf mcp-ping` ✅

**功能**:
- 检测 MCP Chrome Server 是否正确安装
- 测试与浏览器的连接状态
- 显示可用的 MCP 工具列表
- 提供安装提示（如果未安装）

**文件**: `cmd/mcp-ping.go`

**已完成**:
- [x] 实现 mcp-ping 命令
- [x] 检测 MCP 服务器连接
- [x] 列出可用的 Chrome 工具
- [x] 超时处理（5-10秒）
- [x] 跨平台兼容性
- [x] 提供安装提示

---

### 阶段 3：核心功能迁移 🔄

#### 3.1 登录模块简化 ✅

**已删除**:
- `client/Login()` 函数
- `client/ConfigLogin()` 函数
- `cmd/executeWithLoginRetry()` 重试逻辑
- login 配置选项

**保留功能**:
- `client/extractHandleFromProfile()` - 从 profile 页面提取用户名
- `client/extractEmailFromProfile()` - 从 profile 页面提取邮箱

**新设计**:
- 浏览器默认已登录，无需处理登录逻辑
- 用户在浏览器中管理登录状态
- CF-Tool 直接使用浏览器的 cookies

#### 3.2 提交模块重构 ✅

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

#### 3.3 解析模块重构 ✅

**修改文件**: `client/parse.go`, `client/statis.go`

**核心改进**:
1. **Fetcher 抽象**:
```go
// 使用 Fetcher 接口统一数据获取
body, err := c.fetcher.Get(URL)
```

2. **HTML 解析增强**:
```go
// 新增 extractTextContent() 函数
// 支持从嵌套 HTML 中提取纯文本
// 自动去除 HTML 标签，只保留文本内容
```

3. **样本提取修复**:
- 支持新版 Codeforces HTML 结构
- 处理嵌套 div 的情况
- 提取 `<pre>` 标签内的纯文本

**已完成**:
- [x] Fetcher 接口统一
- [x] HTML 内容获取修复
- [x] 样本提取逻辑增强
- [x] 添加调试日志

#### 3.4 其他模块重构 ✅

**涉及文件**:
- `client/watch.go` - 监控提交状态
- `client/pull.go` - 拉取代码
- `client/clone.go` - 克隆用户提交
- `client/statis.go` - 获取比赛统计
- `client/race.go` - 比赛倒计时

**重构策略**:
```go
// 使用 Fetcher 接口统一数据获取
body, err := c.fetcher.Get(url)
```

**已完成**:
- [x] client/watch.go - 移除登录检查，使用 Fetcher
- [x] client/pull.go - 移除登录检查，使用 Fetcher
- [x] client/clone.go - 移除登录检查，使用 Fetcher
- [x] client/statis.go - 移除登录检查，添加日志，使用 Fetcher
- [x] client/race.go - 移除登录检查，使用 Fetcher
- [x] 所有模块统一使用 Fetcher 接口
- [x] 移除所有 findHandle() 登录状态检查

---

### 阶段 3.5：日志系统改进 ✅

#### CF_DEBUG 多级日志支持

**功能**:
- 支持多个日志级别（Debug/Info/Warning/Error）
- 通过环境变量 CF_DEBUG 控制日志详细程度
- 彩色日志输出
- 结构化日志支持

**使用方法**:
```bash
# 详细日志（包含所有调试信息）
CF_DEBUG=debug ./bin/cf parse 100
# 或
CF_DEBUG=1 ./bin/cf parse 100

# 标准日志（只显示重要信息）
CF_DEBUG=info ./bin/cf parse 100
# 或
CF_DEBUG=2 ./bin/cf parse 100
```

**实现文件**:
- `pkg/logger/` - 日志系统实现
- `client/client.go` - CF_DEBUG 环境变量处理

**已完成**:
- [x] 实现分级日志系统
- [x] CF_DEBUG=debug/1 → DebugLevel
- [x] CF_DEBUG=info/2 → InfoLevel
- [x] 彩色日志输出
- [x] 转换 happy path 输出到 logger.Info
- [x] 保留用户重要信息（info.Hint()）的 color.Cyan

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

### 已完成 ✅
- [x] **阶段 1**: Go MCP 客户端库 (`pkg/mcp/`)
- [x] **阶段 1**: 浏览器 HTTP 客户端 (`client/fetcher.go`)
- [x] **阶段 2**: `cf mcp-ping` 测试命令 (`cmd/mcp-ping.go`)
- [x] **阶段 3**: 浏览器模式提交模块 (`client/submit_browser.go`)
- [x] **阶段 3**: 浏览器模式解析模块 (`client/parse.go`)
- [x] **阶段 3**: 其他模块适配（watch, pull, clone, statis, race）
- [x] **功能改进**: 删除 login/logout 功能，简化为浏览器模式
- [x] **功能改进**: 移除所有登录状态检查
- [x] **功能改进**: 优化日志输出，添加 CF_DEBUG 多级支持
- [x] **Bug修复**: 修复 parse 命令样本提取（支持嵌套 HTML 结构）

### 待完成 ⏳
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

- [x] **M1**: 基础设施完成（MCP 客户端 + 浏览器客户端） ✅
- [x] **M2**: `cf mcp-ping` 命令可用 ✅
- [x] **M3**: 功能简化完成（删除登录流程，浏览器默认已登录） ✅
- [x] **M4**: 核心功能迁移完成（parse/submit/watch/pull/clone） ✅
- [x] **M5**: Bug 修复（parse 命令 HTML 提取） ✅
- [x] **M6**: 日志系统改进（CF_DEBUG 多级支持） ✅
- [x] **M7**: Bug 修复完整（新格式 HTML 解析 + 去除多余空行） ✅
- [x] **M8**: 用户文档更新（README.md + README_zh_CN.md） ✅
- [ ] **M9**: 配置文件和安装脚本完成 ⏳
- [ ] **M10**: 发布正式版本 ⏳

---

## 📝 开发日志

### 2025-12-31

#### 浏览器模式核心功能 ✅
- ✅ **MCP 客户端库**: 完成 `pkg/mcp/` 基础设施
  - 支持 stdio 和 HTTP 传输协议
  - 实现 JSON-RPC 2.0 通信
  - 封装常用 Chrome 工具调用

- ✅ **Fetcher 抽象层**: 完成 `client/fetcher.go`
  - 统一的 Fetcher 接口
  - HTTPFetcher（传统模式）
  - BrowserFetcher（浏览器模式）
  - 自动检测并切换模式

- ✅ **核心命令适配**:
  - `cf mcp-ping` - 测试 MCP 连接
  - `cf parse` - 解析题目样本（修复 HTML 提取）
  - `cf submit` - 浏览器模式提交
  - `cf watch` - 监控提交状态
  - `cf pull` - 拉取代码
  - `cf clone` - 克隆用户提交
  - `cf statis` - 获取比赛统计
  - `cf race` - 比赛倒计时

#### 功能简化 ✅
- ✅ **删除登录流程**:
  - 移除 `Login()` 和 `ConfigLogin()` 函数
  - 移除 `executeWithLoginRetry()` 重试逻辑
  - 移除 login 配置选项
  - 简化为：浏览器默认已登录，直接使用

- ✅ **移除登录检查**:
  - 删除 `findHandle()` 登录状态验证
  - 从 parse/submit/watch/statis/clone/race 中移除检查
  - 简化代码逻辑

#### 日志系统改进 ✅
- ✅ **多级日志支持**:
  - 实现 `pkg/logger/` 分级日志系统
  - 支持 Debug/Info/Warning/Error 级别
  - 添加 CF_DEBUG 环境变量支持：
    - `CF_DEBUG=debug` 或 `1` → 详细日志
    - `CF_DEBUG=info` 或 `2` → 标准日志
  - 转换 happy path 的 color.Cyan/Yellow 为 logger.Info

#### Bug 修复 ✅
- ✅ **Parse 命令修复** (Commit: `60c965c`):
  - 修复 BrowserFetcher 获取 HTML 内容
  - 修复样本提取逻辑（支持嵌套 HTML）
  - 添加 `extractTextContent()` 清理 HTML 标签
  - 成功解析 Codeforces 新版 HTML 结构

- ✅ **日志级别改进** (Commit: `701b467`):
  - CF_DEBUG 支持多档位
  - 用户可选择日志详细程度

#### 技术亮点 🌟
- **Fetcher 模式**: 统一接口，HTTP 和浏览器模式无缝切换
- **自动检测**: 启动时自动检测 MCP 服务器并启用浏览器模式
- **HTML 解析增强**: 支持多种 HTML 结构，去除标签提取纯文本
- **零配置**: 用户只需安装 MCP Chrome Server，无需手动配置

#### 待办事项 📋
- [x] HTML 解析完整修复（支持新旧两种格式）
- [x] README 文档更新（添加 mcp-chrome 安装指南）
- [ ] 配置文件格式更新
- [ ] 安装脚本开发
- [ ] 集成测试

#### Hotfix 修复记录 🐛

**Hotfix #1**: HTML 解析完整修复 (2025-12-31 20:30)
- ✅ **问题**: 发现 Codeforces 使用两种 HTML 格式
  - 旧格式: 使用 `<br />` 标签换行
  - 新格式: 使用 `<div class="test-example-line">` 标签分隔
- ✅ **修复**:
  - 添加 `</div>` 标签处理（新格式）
  - 保留 `<br>` 标签处理（旧格式）
  - 添加详细的 INFO 级别日志
  - 测试两种格式都正常工作
- ✅ **提交**: Commit `c25d63d` - HOTFIX - Handle <div> tags in HTML parser for new Codeforces format

**Hotfix #2**: 去除多余换行 (2025-12-31 20:32)
- ✅ **问题**: `</div>` 替换为 `\n` 后，文件末尾多一个空行
- ✅ **修复**: 在 `extractTextContent()` 中 Trim 掉末尾换行符
- ✅ **提交**: Commit `0d4dc45` - HOTFIX - Remove trailing newline to avoid double newlines in test files

**Hotfix #3**: 提交时题目 ID 大写转换 (2025-12-31)
- ✅ **问题**: Codeforces 提交表单要求大写字母（A/B/C），但用户可能输入小写（a/b/c）
- ✅ **修复**: 添加 `strings.ToUpper(problemID)` 转换
- ✅ **提交**: Commit `8667beb` - HOTFIX - Add uppercase conversion for problemID in browser submit

#### 用户文档更新 ✅ (2025-12-31 20:35)
- ✅ **README.md**: 添加 "Browser Mode (Recommended)" 章节
  - mcp-chrome 扩展安装步骤
  - mcp-chrome-bridge 安装命令
  - 验证安装方法（`cf mcp-ping` + `cf mocka`）
  - 强调浏览器模式是必需的
- ✅ **README_zh_CN.md**: 同步中文翻译
  - 语言一致性
  - 术语准确性
- ✅ **包含的关键信息**:
  - Chrome 网上应用店链接
  - npm/yarn 安装命令
  - 默认端口 `http://127.0.0.1:12306/mcp`
  - `cf mcp-ping` 验证步骤
  - `cf mocka` 浏览器测试步骤
  - 新版本必须使用浏览器模式的说明

---

## 📊 今日完整工作总结 (2025-12-31)

### 会话目标 ✅
完成 cf-tool 浏览器模式的最小可行产品（MVP），实现 `parse` 和 `submit` 两个核心命令的完整功能。

### 完成的工作流程

#### 1. 问题发现与分析 (19:00-20:20)
- ✅ 发现 HTML 解析问题：多行输入被合并成一行
- ✅ 调查发现 Codeforces 使用两种 HTML 格式：
  - 旧格式（Contest 1000）: `<br />` 标签
  - 新格式（Contest 2122）: `<div class="test-example-line">` 标签
- ✅ 创建 Hotfix.md 记录问题和解决方案

#### 2. 第一轮修复 (20:15-20:20)
- ✅ 修复 `<br>` 标签处理（旧格式）
- ✅ 测试 1000a 成功
- ✅ 发现新格式问题：2122d 仍然失败

#### 3. 第二轮修复 (20:20-20:30)
- ✅ 添加 `<div>` 标签处理（新格式）
- ✅ 添加详细的 INFO 级别日志
- ✅ 测试两种格式都成功
- ✅ 移除 parse.go 调试代码

#### 4. 第三轮修复 (20:30-20:35)
- ✅ 去除多余换行符（`</div>` 替换导致的末尾空行）
- ✅ 确保文件结尾只有一个 `\n`

#### 5. 提交功能完善 (19:46-20:10)
- ✅ 添加 problemID 大写转换（a → A）
- ✅ 选择正确的提交按钮（`#singlePageSubmitButton`）
- ✅ 测试提交流程完整成功

#### 6. 文档更新 (20:35-21:00)
- ✅ 更新 Hotfix.md 完整记录修复过程
- ✅ 更新 README.md 添加浏览器模式安装指南
- ✅ 更新 README_zh_CN.md 同步中文翻译
- ✅ 更新 TODO.md 记录所有进度

#### 7. 年度公告准备 (21:00-21:30)
- ✅ 创建新年公告（happynewyear.md）
- ✅ 创建无表情符号版本（happynewyear-noemoji.md）
- ✅ 准备发布到社区

### 关键提交记录

```bash
c25d63d - HOTFIX - Handle <div> tags in HTML parser for new format
0d4dc45 - HOTFIX - Remove trailing newline to avoid double newlines
8667beb - HOTFIX - Add uppercase conversion for problemID
503b6a2 - HOTFIX - Handle <br> tags in HTML parser
2929692 - docs: Add mcp-chrome installation guide to README
```

### 技术实现细节

#### HTML 解析核心算法 (`client/html/parser.go`)
```go
func extractTextContent(htmlBytes []byte) string {
    text := string(htmlBytes)

    // STEP 1: Handle <div> tags (NEW format)
    divReg := regexp.MustCompile(`</div>`)
    text = divReg.ReplaceAllString(text, "\n")

    // STEP 2: Handle <br> tags (OLD format)
    brReg := regexp.MustCompile(`<br\s*/?>`)
    text = brReg.ReplaceAllString(text, "\n")

    // STEP 3-7: Clean up, normalize, trim
    // ...

    // STEP 8: Remove trailing newlines
    text = strings.TrimRight(text, "\n\r")

    return text
}
```

**关键设计决策**:
- 按顺序处理：先 `</div>` 后 `<br>`（避免冲突）
- 保留所有换行符：使用 `[ \t]+` 而不是 `\s+`
- 逐行 trim：保留多行结构
- 末尾清理：避免双重换行

#### 提交流程 (`client/browser/submit.go`)
```go
// Step 1: Navigate to submit page
mcpClient.Navigate(ctx, URL)

// Step 2: Select problem (uppercase conversion)
problemIDUpper := strings.ToUpper(problemID)
mcpClient.Fill(ctx, "[name='submittedProblemIndex']", problemIDUpper)

// Step 3: Select language
mcpClient.Fill(ctx, "#programTypeId", langID)

// Step 4: Inject source code via JavaScript
mcpClient.CallTool(ctx, "chrome_javascript", {...})

// Step 5: Click submit button
mcpClient.Click(ctx, "#singlePageSubmitButton")
```

**关键设计决策**:
- 大写转换：Codeforces 表单要求大写字母
- 使用正确的按钮 ID：`#singlePageSubmitButton`
- JavaScript 注入代码：避免字符转义问题

### 测试验证

#### 旧格式测试 (Contest 1000, Problem A)
```bash
$ cf parse 1000 a
✅ Extracted 3 samples
✅ in1.txt: 8 lines, all with proper newlines
```

#### 新格式测试 (Contest 2122, Problem D)
```bash
$ cf parse 2122 d
✅ Extracted 1 sample
✅ in1.txt: 13 lines, all properly separated (was 1 line before)
```

#### 提交测试
```bash
$ cf submit 101 a
✅ Selecting problem: a (converted to: A)
✅ Code submitted successfully via browser
✅ Submission ID=355976655, problem=A - Homework
```

### 已知问题和限制

#### 当前限制
1. **配置文件**: 仍使用旧的配置格式（未迁移到浏览器模式）
2. **安装脚本**: 需要用户手动安装 mcp-chrome-bridge
3. **错误处理**: 浏览器通信失败时的错误提示不够友好
4. **依赖项**: 必须运行 mcp-chrome-bridge 后才能使用 cf-tool

#### 不影响核心功能的问题
- WatchSubmission 监控有时需要多次重试才能找到最新提交（非阻塞）
- 某些边缘情况的 HTML 结构可能需要进一步调整

### 待办事项（优先级排序）

#### 高优先级（必须完成）
1. **配置文件迁移**
   - 移除 login 相关配置项
   - 添加 MCP 服务器地址配置（默认：`http://127.0.0.1:12306/mcp`）
   - 更新 `cf config` 命令

2. **错误处理改进**
   - MCP 服务器未启动时的友好提示
   - 浏览器未打开时的自动提示
   - 网络错误重试机制

#### 中优先级（改进体验）
3. **安装脚本**
   - 自动检测并安装 mcp-chrome-bridge
   - 验证安装并提示用户
   - 一键安装脚本

4. **测试增强**
   - 单元测试（MCP 客户端）
   - 集成测试（完整工作流）
   - 跨平台测试（Windows/Linux/macOS）

#### 低优先级（未来功能）
5. **功能增强**
   - 并发解析多个题目
   - 缓存机制
   - 配额管理

6. **文档完善**
   - 视频教程
   - 故障排查指南
   - API 文档

### 文件清单

#### 新增文件
- `Hotfix.md` - 完整的 Bug 修复记录
- `Future.md` - 未来架构设计文档
- `happynewyear.md` - 新年公告（含 emoji）
- `happynewyear-noemoji.md` - 新年公告（无 emoji）

#### 修改的核心文件
- `client/html/parser.go` - HTML 解析（支持新旧格式）
- `client/browser/submit.go` - 浏览器模式提交
- `client/parse.go` - 解析命令入口
- `client/submit.go` - 提交命令入口
- `pkg/logger/` - 分级日志系统
- `pkg/mcp/` - MCP 客户端库
- `README.md` - 英文文档（添加浏览器模式）
- `README_zh_CN.md` - 中文文档（同步更新）
- `TODO.md` - 项目进度追踪

### 会话重启指南

#### 快速恢复上下文
当重启 Claude Code 会话时，需要了解的关键信息：

1. **项目状态**: 浏览器模式 MVP 完成，parse 和 submit 可用
2. **核心功能**: 已实现并测试通过
3. **当前分支**: `chrome-mcp`（功能分支）
4. **主分支**: `master`（稳定版本）
5. **下一步**: 配置文件迁移和错误处理改进

#### 快速测试命令
```bash
# 验证 MCP 连接
cf mcp-ping

# 测试浏览器自动化
cf mocka

# 测试解析（旧格式）
cf parse 1000 a

# 测试解析（新格式）
cf parse 2122 d

# 测试提交
cd cf/contest/101/a && cf submit
```

#### 重要参考文档
- `Hotfix.md` - Bug 修复历史（如遇问题先查阅）
- `Future.md` - 架构设计未来方向
- `TODO.md` - 本文件，完整进度追踪

### 技术债务
1. **代码清理**: 移除所有 `// TODO` 和 `// FIXME` 注释
2. **测试覆盖**: 当前没有单元测试，需要补充
3. **文档完善**: 部分命令缺少详细文档
4. **性能优化**: 可以考虑并发处理多个题目

### 社区准备
- ✅ 年新公告已准备（两个版本）
- ✅ README 已更新（安装指南完整）
- ✅ 核心功能已验证（parse + submit 正常工作）
- ⏳ 待发布：准备发布到社区

---

## 🔗 参考资源

- [MCP-Chrome GitHub](https://github.com/hangwin/mcp-chrome)
- [MCP Protocol Spec](https://modelcontextprotocol.io/)
- [Chrome Extension Docs](https://developer.chrome.com/docs/extensions/)
- [CF-Tool 原项目](https://github.com/xalanq/cf-tool)

---

**最后更新**: 2025-12-31
**维护者**: @NetWilliam
