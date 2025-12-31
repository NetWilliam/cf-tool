# CF-Tool 未来重构计划

> **文档创建日期**: 2025-12-31
> **当前状态**: 浏览器模式基本完成，等待完整的 Delegator 架构重构
> **维护者**: @NetWilliam

---

## 📋 目录

1. [当前架构总结](#当前架构总结)
2. [完整的 Delegator 架构设计](#完整的-delegator-架构设计)
3. [重构实施计划](#重构实施计划)
4. [技术细节](#技术细节)
5. [测试策略](#测试策略)

---

## 🎯 当前架构总结

### 已完成的工作

#### 1. 浏览器模式 ✅
- ✅ MCP 客户端库 (`pkg/mcp/`)
- ✅ Fetcher 接口 (`client/fetcher.go`)
  - `HTTPFetcher` - 传统 HTTP 模式
  - `BrowserFetcher` - 浏览器模式
- ✅ `cf mcp-ping` 测试命令
- ✅ 浏览器自动化提交 (`client/submit.go`)

#### 2. 核心命令迁移 ✅
- ✅ `cf parse` - 使用 fetcher 接口
- ✅ `cf submit` - 使用浏览器自动化（绕过 CSRF）
- ✅ `cf watch` - 使用 fetcher 接口
- ✅ `cf pull`, `cf clone`, `cf race` - 使用 fetcher 接口

#### 3. 功能简化 ✅
- ✅ 删除登录流程，浏览器默认已登录
- ✅ 移除所有 `findHandle()` 登录检查
- ✅ 添加 CF_DEBUG 多级日志支持

### 当前架构问题

#### 问题 1: HTML 解析逻辑散落各处
```go
// client/parse.go - HTML 解析逻辑
func findSample(body []byte) (input, output [][]byte, err error) {
    inputReg := regexp.MustCompile(`<div[^>]*class="input"[^>]*>...`)
    // Codeforces 特定的正则表达式
}

// client/watch.go - 另一组 HTML 解析逻辑
func (c *Client) getSubmissions(URL string, n int) ([]Submission, error) {
    // 更多的 HTML 解析
}
```

**问题**：
- 如果要支持 AtCoder，需要复制这些逻辑到 `client/atcoder/`
- 难以单独测试 HTML 解析逻辑
- HTML 结构变化时，需要修改多个地方

#### 问题 2: 平台特定逻辑与通用逻辑混在一起
```go
// client/parse.go 既包含通用的文件操作，又包含 Codeforces 特定的解析
func (c *Client) ParseProblem(URL, path string, mu *sync.Mutex) (samples int, standardIO bool, err error) {
    body, err := c.fetcher.Get(URL)  // 通用
    input, output, err := findSample(body)  // Codeforces 特定
    for i := 0; i < len(input); i++ {  // 通用
        fileIn := filepath.Join(path, fmt.Sprintf("in%v.txt", i+1))  // 通用
        os.WriteFile(fileIn, input[i], 0644)  // 通用
    }
}
```

#### 问题 3: 循环依赖阻碍重构
```
client/ → client/codeforces/ → client/ (循环！)
```

---

## 🏗️ 完整的 Delegator 架构设计

### 设计目标

1. **平台抽象**: 所有平台特定的 HTML/JSON/JS 逻辑封装在平台包中
2. **接口统一**: command 层只调用统一的 Delegator 接口
3. **易于测试**: 每个平台包可以独立测试
4. **易于扩展**: 添加新平台只需实现 Delegator 接口

### 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      Command Layer                          │
│                                                               │
│  cmd/submit.go  ──→  cln.Submit(info, lang, source)         │
│  cmd/parse.go   ──→  cln.Parse(info)                        │
│  cmd/watch.go   ──→  cln.Watch(info, count)                 │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      Client Layer                           │
│                                                               │
│  client/                                                     │
│    ├── client.go           # Client 结构体                   │
│    ├── info.go             # Info 结构体（移到 pkg/types/）  │
│    └── lang.go             # 语言配置                        │
└──────────────────────────┬──────────────────────────────────┘
                           │ uses Delegator interface
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   Types & Interfaces                       │
│                                                               │
│  pkg/types/                                                  │
│    ├── info.go             # Info 结构体定义                 │
│    ├── delegator.go        # Delegator 接口定义              │
│    └── testcase.go         # Testcase, Submission 结构体     │
└──────────────────────────┬──────────────────────────────────┘
                           │ implements
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Platform Packages                        │
│                                                               │
│  client/codeforces/          client/atcoder/                 │
│    ├── codeforces.go            # atcoder.go                │
│    ├── html_parser.go           # html_parser.go            │
│    ├── submit.go                # submit.go                 │
│    └── watch.go                 # watch.go                  │
└─────────────────────────────────────────────────────────────┘
```

### 核心接口定义

```go
// pkg/types/delegator.go
package types

import "context"

// Info 包含题目和提交信息
type Info struct {
    ProblemType  string // "contest", "gym", "problemset"
    GroupID      string
    ContestID    string
    ProblemID    string
    SubmissionID string
}

// Testcase 测试用例
type Testcase struct {
    Input  string
    Output string
}

// Submission 提交记录
type Submission struct {
    ID              string
    ProblemID       string
    ContestID       string
    Lang            string
    Status          string
    Time            string
    Memory          string
    PassedTestCount int
    IsFinal         bool
}

// Delegator 平台代理接口
type Delegator interface {
    // ParseProblem 解析题目页面，提取测试样例
    ParseProblem(ctx context.Context, info Info) ([]Testcase, error)

    // SubmitCode 提交代码到平台
    SubmitCode(ctx context.Context, info Info, langID, source string) (submissionID string, err error)

    // WatchSubmission 监控提交状态
    WatchSubmission(ctx context.Context, info Info, count int) ([]Submission, error)

    // PullCode 拉取源代码
    PullCode(ctx context.Context, submissionID string) (source string, lang string, err error)

    // GetProblemURL 获取题目 URL
    GetProblemURL(info Info) (string, error)

    // GetSubmitURL 获取提交 URL
    GetSubmitURL(info Info) (string, error)

    // GetMySubmissionsURL 获取我的提交列表 URL
    GetMySubmissionsURL(info Info) (string, error)
}
```

### 平台包结构

```go
// client/codeforces/html_parser.go
package codeforces

import "github.com/NetWilliam/cf-tool/pkg/types"

// ParseTestcasesFromHTML 从 Codeforces HTML 提取测试样例
func ParseTestcasesFromHTML(html string) ([]types.Testcase, error) {
    // Codeforces 特定的 HTML 解析逻辑
    inputReg := regexp.MustCompile(`<div[^>]*class="input"[^>]*>...`)
    // ...
}

// ExtractSubmissionsFromHTML 从 Codeforces HTML 提取提交记录
func ExtractSubmissionsFromHTML(html string) ([]types.Submission, error) {
    // Codeforces 特定的解析逻辑
}
```

```go
// client/codeforces/codeforces.go
package codeforces

import (
    "context"
    "github.com/NetWilliam/cf-tool/pkg/types"
    "github.com/NetWilliam/cf-tool/pkg/mcp"
)

// CodeforcesDelegator 实现 Delegator 接口
type CodeforcesDelegator struct {
    mcpClient *mcp.Client
    host      string
}

func NewCodeforcesDelegator(mcpClient *mcp.Client, host string) types.Delegator {
    return &CodeforcesDelegator{mcpClient: mcpClient, host: host}
}

func (d *CodeforcesDelegator) ParseProblem(ctx context.Context, info types.Info) ([]types.Testcase, error) {
    // 使用 ParseTestcasesFromHTML
    html := d.fetchPage(info)
    return ParseTestcasesFromHTML(html)
}

// ... 其他接口实现
```

---

## 🔄 重构实施计划

### 阶段 1: 解决循环依赖 ✅ (部分完成)

#### 1.1 创建共享类型包
- ✅ 创建 `pkg/types/delegator.go` - Delegator 接口
- ✅ 创建 `pkg/types/info.go` - Info 结构体
- ⏳ 将 `client/info.go` 中的 `Info` 移到 `pkg/types/info.go`
- ⏳ 更新所有引用 `client.Info` 的地方改为 `types.Info`

#### 1.2 需要修改的文件（20+ 个）
```
client/info.go           → 移到 pkg/types/info.go
client/*.go              → 更新 import 和类型引用
cmd/*.go                 → 更新 import 和类型引用
config/*.go              → 更新 import 和类型引用
```

**工作量估算**: 2-3 小时

### 阶段 2: 创建平台包

#### 2.1 Codeforces 平台包
- ✅ 创建 `client/codeforces/codeforces.go`
- ⏳ 创建 `client/codeforces/html_parser.go`
- ⏳ 创建 `client/codeforces/submit.go`
- ⏳ 创建 `client/codeforces/watch.go`

**功能划分**:
```
codeforces.go          - 实现 Delegator 接口，协调其他模块
html_parser.go         - HTML 解析（findSample, findSubmission 等）
submit.go              - 浏览器自动化提交逻辑
watch.go               - 提交监控逻辑
```

**工作量估算**: 3-4 小时

#### 2.2 AtCoder 平台包（未来）
```
client/atcoder/
  ├── atcoder.go
  ├── html_parser.go
  ├── submit.go
  └── watch.go
```

**工作量估算**: 4-5 小时

### 阶段 3: 重构 client 层

#### 3.1 简化 Client 结构体
```go
// client/client.go
type Client struct {
    // ... 现有字段

    // 新增字段
    delegator types.Delegator  // 平台代理（Codeforces, AtCoder 等）
}
```

#### 3.2 更新初始化逻辑
```go
// client/client.go
func (c *Client) initBrowserMode() error {
    // ... 现有逻辑

    // 根据平台创建对应的 delegator
    if strings.Contains(c.host, "codeforces.com") {
        c.delegator = codeforces.NewCodeforcesDelegator(c.mcpClient, c.host)
    } else if strings.Contains(c.host, "atcoder.jp") {
        c.delegator = atcoder.NewAtCoderDelegator(c.mcpClient, c.host)
    }
}
```

#### 3.3 重构命令方法
```go
// client/parse.go
func (c *Client) Parse(info types.Info) ([]string, []string, error) {
    testcases, err := c.delegator.ParseProblem(context.Background(), info)
    // 保存测试样例到文件（通用逻辑）
    // ...
}

// client/submit.go
func (c *Client) Submit(info types.Info, langID, source string) error {
    submissionID, err := c.delegator.SubmitCode(context.Background(), info, langID, source)
    // 监控提交（通用逻辑）
    // ...
}

// client/watch.go
func (c *Client) WatchSubmission(info types.Info, count int) ([]types.Submission, error) {
    return c.delegator.WatchSubmission(context.Background(), info, count)
}
```

**工作量估算**: 4-5 小时

### 阶段 4: 测试和验证

#### 4.1 单元测试
```
client/codeforces/html_parser_test.go
client/atcoder/html_parser_test.go
```

#### 4.2 集成测试
```
scripts/test-parse.sh
scripts/test-submit.sh
scripts/test-watch.sh
```

**工作量估算**: 2-3 小时

---

## 🔧 技术细节

### HTML 解析策略

#### Codeforces
```go
// 测试样例结构
<div class="input">
    <div class="title">Input</div>
    <pre>...</pre>
</div>

// 提交表格结构
<form id="submitForm">
    <input name="csrf_token" />
    <select name="programTypeId" />
    <textarea name="source"></textarea>
</form>
```

#### AtCoder
```go
// 测试样例结构（不同！）
<div class="io-style">
    <div class="part">
        <section>
            <h3>入力例 1</h3>
            <pre>...</pre>
        </section>
    </div>
</div>

// 提交需要 CSRF token 和不同的字段名
```

### 浏览器自动化差异

#### Codeforces
```javascript
// 选择语言
document.querySelector('[name="programTypeId"]').value = "54";

// 设置代码
document.querySelector('[name="source"]').value = source;

// 提交
document.querySelector('input[type="submit"]').click();
```

#### AtCoder
```javascript
// 选择语言
document.querySelector('#select-lang').value = "4003";

// 设置代码
document.querySelector('#source-code').value = source;

// 提交（不同的按钮！）
document.querySelector('.btn-submit').click();
```

---

## 🧪 测试策略

### 单元测试

```go
// client/codeforces/html_parser_test.go
package codeforces_test

func TestParseTestcasesFromHTML(t *testing.T) {
    html := `
        <div class="input"><pre>1 2</pre></div>
        <div class="output"><pre>3</pre></div>
    `
    testcases, err := ParseTestcasesFromHTML(html)
    assert.NoError(t, err)
    assert.Equal(t, 1, len(testcases))
    assert.Equal(t, "1 2", testcases[0].Input)
    assert.Equal(t, "3", testcases[0].Output)
}
```

### 集成测试

```bash
#!/bin/bash
# scripts/test-parse.sh

echo "Testing parse command..."
cf parse 100 A
if [ -f "./cf/contest/100/a/in1.txt" ]; then
    echo "✓ Parse test passed"
else
    echo "✗ Parse test failed"
    exit 1
fi
```

---

## 📊 工作量估算

| 阶段 | 任务 | 工作量 | 状态 |
|------|------|--------|------|
| 1 | 解决循环依赖 | 2-3h | ⏳ 50% |
| 2 | 创建 Codeforces 平台包 | 3-4h | ✅ 80% |
| 3 | 重构 client 层 | 4-5h | ⏳ 0% |
| 4 | 测试和验证 | 2-3h | ⏳ 0% |
| **总计** | | **11-15h** | **⏳ 30%** |

---

## 🎯 优先级建议

### P0 - 核心功能（必须完成）
1. ✅ 解决循环依赖（Info 移到 pkg/types）
2. ✅ Codeforces 平台包基本功能
3. ⏳ Parse/Submit/Watch 使用 Delegator

### P1 - 重要功能
4. ⏳ 单元测试
5. ⏳ 集成测试

### P2 - 增强功能
6. ⏳ AtCoder 平台支持
7. ⏳ 错误处理改进
8. ⏳ 性能优化

---

## 📚 参考资料

### 相关文档
- [MCP Chrome Server](https://github.com/hangwin/mcp-chrome)
- [MCP Protocol Spec](https://modelcontextprotocol.io/)
- [Codeforces API](https://codeforces.com/apiHelp)

### 设计模式
- **Proxy Pattern**: Delegator 作为平台代理
- **Strategy Pattern**: 不同平台实现相同接口
- **Factory Pattern**: 根据域名创建对应 Delegator

---

## 🔄 迭代计划

### 当前迭代 (v1.0)
- 目标: 完成基本的浏览器模式和简化的 Delegator
- 状态: 进行中
- 预计完成: 2025-12-31

### 下一迭代 (v2.0)
- 目标: 完整的 Delegator 架构
- 预计开始: 2026-01-01
- 预计完成: 2-3 周后

### 未来迭代 (v3.0+)
- 多平台支持 (AtCoder, CodeChef, etc.)
- 性能优化
- UI 改进

---

**最后更新**: 2025-12-31
**下次审查**: 2026-01-01
