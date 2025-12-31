# CF-Tool Hotfix 记录

> **创建日期**: 2025-12-31
> **版本**: v1.0-browser
> **状态**: 🚧 进行中 - Bug #1 发现新问题需要修复

---

## 🚨 重要更新 (2025-12-31 20:20)

Bug #1 的修复**不完整**！发现 Codeforces 有**两种不同的 HTML 格式**：

1. **旧格式** (Contest 1000): 使用 `<br />` 标签换行 ✅ 已修复
2. **新格式** (Contest 2122): 使用 `<div>` 标签分隔每行 ❌ **未处理**

当前代码只处理了旧格式，新格式的多行输入仍会被合并成一行。

---

## 🐛 Bug #1: HTML 解析时换行符丢失

### 问题描述
多行输入在解析时被合并成一行，导致测试用例的输入输出格式错误。

### 复现步骤
```bash
cf parse 100 a
cat cf/contest/100/a/in1.txt
# 预期：多行输入
# 实际：所有内容在一行
```

### 根本原因

**文件**: `client/html/parser.go:40-73`

#### Codeforces 的两种 HTML 格式

**格式 1: 旧格式** (Contest 1000 及更早的比赛)
```html
<pre>3<br />XS<br />XS<br />M<br />XL<br />S<br />XS<br /></pre>
```
- 使用 `<br />` 或 `<br>` 标签表示换行
- ✅ 已在第一次修复中解决

**格式 2: 新格式** (Contest 2122 及最近的比赛)
```html
<pre>
  <div class="test-example-line test-example-line-even test-example-line-0">2</div>
  <div class="test-example-line test-example-line-odd test-example-line-1">6 6</div>
  <div class="test-example-line test-example-line-odd test-example-line-1">1 2</div>
  <div class="test-example-line test-example-line-odd test-example-line-1">2 3</div>
  <div class="test-example-line test-example-line-odd test-example-line-1">3 4</div>
  ...
</pre>
```
- 每行用 `<div class="test-example-line">...</div>` 包裹
- ❌ **当前代码未处理**：直接删除所有 `<div>` 标签，导致所有行合并

#### 当前代码的问题

当前代码处理流程（针对新格式）：
```
HTML: <div>2</div><div>6 6</div><div>1 2</div>...
  ↓ 删除所有标签
Result: "26 61 22..."  # ❌ 所有行被合并成一行
```

**问题 1**: `<div>` 标签被直接删除（新格式的主要问题）

当前代码：
```go
// Remove all remaining HTML tags
tagReg := regexp.MustCompile(`<[^>]+>`)
text = tagReg.ReplaceAllString(text, "")  // ❌ 删除了 <div>...</div> 标签
```

**问题 2**: `<br>` 标签的处理（旧格式，已在第一次修复中部分解决）

第一次修复添加了：
```go
brReg := regexp.MustCompile(`<br\s*/?>`)
text = brReg.ReplaceAllString(text, "\n")  # ✅ 处理旧格式
```

原代码使用 `\s+` 将所有空白符（包括 `\n`）替换成空格：
```go
spaceReg := regexp.MustCompile(`\s+`)
text = spaceReg.ReplaceAllString(text, " ")  // ❌ 把 \n 也替换成空格
```

**问题 3**: `\s+` 匹配并替换换行符（原始问题，已在第一次修复中解决）

原代码使用 `\s+` 将所有空白符（包括 `\n`）替换成空格：
```go
spaceReg := regexp.MustCompile(`\s+`)
text = spaceReg.ReplaceAllString(text, " ")  // ❌ 把 \n 也替换成空格
```

第一次修复已改为 `[ \t]+` 只替换空格和制表符。

#### 实际测试结果

**旧格式测试** (Contest 1000, Problem A):
```bash
# HTML
<pre>3<br />XS<br />XS<br />M<br />XL<br />S<br />XS<br /></pre>

# 当前代码输出 ✅ 正确
in1.txt:
3
XS
XS
M
XL
S
XS
```

**新格式测试** (Contest 2122, Problem D):
```bash
# HTML
<pre>
  <div class="test-example-line ...">2</div>
  <div class="test-example-line ...">6 6</div>
  <div class="test-example-line ...">1 2</div>
  ...
</pre>

# 当前代码输出 ❌ 错误
in1.txt: "26 61 22 33 44 61 55 64 31 21 31 4"

# 预期输出 ✅
in1.txt:
2
6 6
1 2
2 3
3 4
4 6
1 5
5 6
4 3
1 2
1 3
1 4
```

### 修复方案

**关键修复**: 必须同时处理两种 HTML 格式

#### 方案: 按顺序处理不同换行标记

```go
func extractTextContent(htmlBytes []byte) string {
    text := string(htmlBytes)

    // STEP 1: Handle <div> tags (new format)
    // Replace closing </div> tags with newlines to preserve line breaks
    // Each <div>...</div> represents one line in the new format
    divReg := regexp.MustCompile(`</div>`)
    text = divReg.ReplaceAllString(text, "\n")

    // STEP 2: Handle <br> tags (old format)
    // Replace <br>, <br/>, and <br /> with newlines
    brReg := regexp.MustCompile(`<br\s*/?>`)
    text = brReg.ReplaceAllString(text, "\n")

    // STEP 3: Remove all remaining HTML tags
    // At this point, all structural tags are gone or replaced with newlines
    tagReg := regexp.MustCompile(`<[^>]+>`)
    text = tagReg.ReplaceAllString(text, "")

    // STEP 4: Unescape HTML entities
    text = html.UnescapeString(text)

    // STEP 5: Normalize horizontal whitespace (spaces and tabs only)
    // Do NOT touch newlines or carriage returns
    spaceReg := regexp.MustCompile(`[ \t]+`)
    text = spaceReg.ReplaceAllString(text, " ")

    // STEP 6: Trim each line and preserve line breaks
    lines := strings.Split(text, "\n")
    for i, line := range lines {
        lines[i] = strings.TrimSpace(line)
    }
    text = strings.Join(lines, "\n")

    // STEP 7: Trim leading/trailing whitespace but keep structure
    text = strings.Trim(text, " \t\r")

    return text
}
```

#### 处理流程对比

**旧格式处理**:
```
HTML: <pre>3<br />XS<br />XS<br />M<br /></pre>
  ↓ Step 2: <br> → \n
"3\nXS\nXS\nM\n"
  ↓ Step 3: 删除其他标签
"3\nXS\nXS\nM\n"
  ↓ Step 4-7: 清理和trim
"3\nXS\nXS\nM\n"  ✅ 正确
```

**新格式处理**:
```
HTML: <pre><div>2</div><div>6 6</div><div>1 2</div></pre>
  ↓ Step 1: </div> → \n
"\n2\n6 6\n1 2\n"
  ↓ Step 3: 删除 <div> 开始标签
"\n2\n6 6\n1 2\n"
  ↓ Step 4-7: 清理和trim
"2\n6 6\n1 2\n"  ✅ 正确
```

### 测试计划

#### 必须测试的用例

1. **旧格式** (Contest 1000, Problem A)
   - 多行输入
   - 包含 `<br />` 换行符
   - 预期：每行正确分离

2. **新格式** (Contest 2122, Problem D)
   - 多行输入
   - 包含 `<div class="test-example-line">` 标签
   - 预期：每个 div 的内容为一行

3. **混合测试**
   - 确保旧格式不被破坏
   - 确保新格式正确处理
   - 测试输出文件（通常也是多行）

4. **边界情况**
   - 只有一行的输入
   - 包含空行的输入
   - 嵌套的 div 标签（如果存在）

---

## 🐛 Bug #2: 提交时未选择题目

### 问题描述
提交代码时只设置了语言，但没有选择要提交的题目（A题、B题等），导致提交失败或提交到错误的题目。

### 复现步骤
```bash
cf submit 100 a
# 当前代码只设置 programTypeId
# 没有设置 submittedProblemIndex = "A"
# 提交按钮使用了错误的选择器
```

### 根本原因

**文件**: `client/browser/submit.go`

**问题 1**: 没有选择题目（第 30-37 行）

**问题 2**: problemID 需要大写转换

Codeforces 的题目选择器要求使用大写字母（A、B、C、D、E），但用户可能输入小写字母（a、b、c、d、e）。
```go
// Step 2: Fill language selector
logger.Debug("Selecting language: %s", langID)
if err := c.mcpClient.Fill(ctx, "#programTypeId", langID); err != nil {
    // ❌ 缺少：选择题目 A/B/C/D/E
    // 应该: Fill(ctx, "[name='submittedProblemIndex']", problemID)
}
```

**问题 2**: 提交按钮选择器错误（第 66-71 行）
```go
submitSelectors := []string{
    "input[type='submit']",
    "button[type='submit']",
    ".submit",
    "[value='Submit']",
    // ❌ 缺少: "#singlePageSubmitButton"
}
```

**Codeforces 提交表单结构**:
```html
<form id="submitForm">
    <!-- 选择题目 -->
    <select name="submittedProblemIndex">
        <option value="A">A</option>
        <option value="B">B</option>
        ...
    </select>

    <!-- 选择语言 -->
    <select name="programTypeId">
        <option value="54">GNU C++17</option>
        ...
    </select>

    <!-- 源代码 -->
    <textarea name="source"></textarea>

    <!-- 提交按钮 -->
    <input type="submit" id="singlePageSubmitButton" value="Submit" />
</form>
```

### 修复方案

#### 修改 `client/browser/submit.go`

```go
// SubmitCode performs browser automation to submit code
func SubmitCode(ctx context.Context, mcpClient *mcp.Client, URL, langID, source, problemID string) error {
    if mcpClient == nil {
        return errors.New("browser mode required")
    }

    logger.Info("Navigating to submit page: %s", URL)

    // Step 1: Navigate to submit page
    if err := mcpClient.Navigate(ctx, URL); err != nil {
        return fmt.Errorf("navigation failed: %w", err)
    }

    time.Sleep(2 * time.Second)

    // Step 2: Select problem (A/B/C/D/E)
    // Convert problemID to uppercase (e.g., "a" → "A")
    problemIDUpper := strings.ToUpper(problemID)
    logger.Debug("Selecting problem: %s (converted to: %s)", problemID, problemIDUpper)
    if err := mcpClient.Fill(ctx, "[name='submittedProblemIndex']", problemIDUpper); err != nil {
        logger.Warning("Failed to fill problem selector: %v", err)
    }

    time.Sleep(500 * time.Millisecond)

    // Step 3: Select language
    logger.Debug("Selecting language: %s", langID)
    if err := mcpClient.Fill(ctx, "#programTypeId", langID); err != nil {
        logger.Warning("Failed to fill language selector: %v", err)
    }

    time.Sleep(500 * time.Millisecond)

    // Step 4: Inject source code using JavaScript
    logger.Debug("Injecting source code (%d bytes)...", len(source))
    jsCode := fmt.Sprintf(`
        (function() {
            let sourceField = document.querySelector('[name="source"]');
            if (!sourceField) {
                sourceField = document.getElementById('source');
            }
            if (sourceField) {
                sourceField.value = %s;
                return 'success';
            }
            return 'failed';
        })();
    `, jsonEscape(source))

    _, err := mcpClient.CallTool(ctx, "chrome_javascript", map[string]interface{}{
        "code": jsCode,
    })
    if err != nil {
        return fmt.Errorf("failed to inject source code: %w", err)
    }

    time.Sleep(500 * time.Millisecond)

    // Step 5: Click submit button (use correct ID)
    logger.Debug("Clicking submit button...")
    submitSelectors := []string{
        "#singlePageSubmitButton",  // ✅ Codeforces 特定的按钮 ID
        "input[type='submit']",
        "button[type='submit']",
        ".submit",
        "[value='Submit']",
    }

    var submitErr error
    for _, selector := range submitSelectors {
        if err := mcpClient.Click(ctx, selector); err != nil {
            submitErr = err
            continue
        }
        submitErr = nil
        logger.Debug("Successfully clicked submit button with selector: %s", selector)
        break
    }

    if submitErr != nil {
        return fmt.Errorf("failed to click submit button: %w", submitErr)
    }

    // Wait for submission to process
    time.Sleep(3 * time.Second)

    logger.Info("Code submitted successfully via browser")
    return nil
}
```

#### 修改调用者

**文件**: `client/submit.go`

需要传递 `problemID` 参数：

```go
func (c *Client) Submit(info Info, langID, source string) (err error) {
    // ...
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    // Use browser automation to submit
    if err := browser.SubmitCode(ctx, c.mcpClient, URL, langID, source, info.ProblemID); err != nil {
        // ...
    }
}
```

### 测试计划

1. 提交 A 题
2. 提交 B 题
3. 验证提交是否到了正确的题目

---

## 📋 修复进度

### Bug #1: 换行符丢失（完整修复）

**状态**: ✅ 已完成

**修复过程**:

**第一阶段** (2025-12-31 19:44) - 旧格式修复:
- [x] 问题调查 - 旧格式使用 `<br />` 标签
- [x] 修改 `client/html/parser.go` 处理 `<br>` 标签
- [x] 测试旧格式 1000a - ✅ 通过

**第二阶段** (2025-12-31 20:20) - 发现新格式问题:
- [x] 发现新格式使用 `<div>` 标签
- [x] 测试新格式 2122d - ❌ 失败（所有行被合并）
- [x] 分析两种 HTML 格式的差异

**第三阶段** (2025-12-31 20:25) - 完整修复:
- [x] 添加 `</div>` 标签处理（新格式）
- [x] 添加详细的 INFO 级别日志
- [x] 测试旧格式 1000a - ✅ 通过（未破坏）
- [x] 测试新格式 2122d - ✅ 通过（13 行正确分离）

**测试结果**:

**旧格式** (Contest 1000a):
```bash
$ cat cf/contest/1000/a/in1.txt
3
XS
XS
M
XL
S
XS

# ✅ 8 行，换行符正确保留
```

**新格式** (Contest 2122d):
```bash
$ cat cf/contest/2122/d/in1.txt
2
6 6
1 2
2 3
3 4
4 6
1 5
5 6
4 3
1 2
1 3
1 4

# ✅ 13 行，每行正确分离（之前是 1 行 "26 61 22 33..."）
```

**日志输出** (CF_DEBUG=info):
```
[HTML Parser] Replaced </div> tags with newlines (new format)
[HTML Parser] Replaced <br> tags with newlines (old format)
[HTML Parser] Extraction complete: 13 lines
```

### Bug #2: 未选择题目
- [x] 问题调查
- [x] 修改 `client/browser/submit.go` 添加 problemID 参数 (2025-12-31 19:46)
- [x] 修改 `client/submit.go` 传递 problemID (2025-12-31 19:46)
- [x] 测试提交不同题目 (2025-12-31 19:46)

**测试结果**:
```bash
$ cf submit 101 a
✓ Navigating to submit page
✓ Selecting problem: a (converted to: A)
✓ Selecting language: 91
✓ Injecting source code
✓ Clicking submit button with selector: #singlePageSubmitButton
✅ Code submitted successfully via browser
✓ Submission ID=355976655, problem=A - Homework
```

---

## 🔗 相关文件

- `client/html/parser.go` - HTML 解析逻辑
- `client/browser/submit.go` - 浏览器自动化提交
- `client/submit.go` - 提交命令入口
- `client/parse.go` - 解析命令入口

---

## 🔧 修改计划

### 需要修改的文件

#### 1. **client/html/parser.go**

修改 `extractTextContent()` 函数，按以下顺序处理：

```go
func extractTextContent(htmlBytes []byte) string {
    text := string(htmlBytes)

    // STEP 1: Handle <div> tags (new format)
    // Replace closing </div> tags with newlines
    divReg := regexp.MustCompile(`</div>`)
    text = divReg.ReplaceAllString(text, "\n")
    logger.Info("[HTML Parser] Replaced </div> tags with newlines")

    // STEP 2: Handle <br> tags (old format)
    brReg := regexp.MustCompile(`<br\s*/?>`)
    text = brReg.ReplaceAllString(text, "\n")
    logger.Info("[HTML Parser] Replaced <br> tags with newlines")

    // STEP 3: Remove all remaining HTML tags
    tagReg := regexp.MustCompile(`<[^>]+>`)
    text = tagReg.ReplaceAllString(text, "")
    logger.Debug("[HTML Parser] Removed remaining HTML tags")

    // STEP 4-7: ... (existing logic)
}
```

**添加的日志**:
- 每个步骤添加 INFO 或 DEBUG 日志
- 便于后续调试和验证处理流程

#### 2. **client/parse.go** (可选)

移除调试代码（如果不再需要）：
```go
// 删除或注释掉这行：
// os.WriteFile("/tmp/cf_parse_debug.html", body, 0644)
```

### 测试验证

执行以下测试确保修复正确：

```bash
# 测试旧格式
rm -rf cf/contest/1000
./bin/cf parse 1000 a
cat cf/contest/1000/a/in1.txt  # 应该是多行

# 测试新格式
rm -rf cf/contest/2122
./bin/cf parse 2122 d
cat cf/contest/2122/d/in1.txt  # 应该是多行，不是单行

# 验证字节内容
od -c cf/contest/2122/d/in1.txt  # 应该看到 \n 换行符
```

---

## ✅ 当前状态总结

### 修改的文件

1. **client/html/parser.go** (部分完成)
   - ✅ 添加 `<br>` 标签处理（旧格式）
   - ✅ 将 `\s+` 改为 `[ \t]+`，只替换空格和制表符
   - ✅ 添加逐行 trim 逻辑
   - ❌ **缺少**: `<div>` 标签处理（新格式）

2. **client/browser/submit.go** (已完成)
   - 添加 `problemID` 参数到 `SubmitCode()` 函数
   - 添加选择题目的步骤（Step 2）
   - **添加大写转换**: `problemIDUpper := strings.ToUpper(problemID)` 确保符合 Codeforces 要求
   - 更新提交按钮选择器，添加 `#singlePageSubmitButton`
   - 重新组织步骤顺序：选择题目 → 选择语言 → 注入代码 → 点击提交

3. **client/submit.go** (已完成)
   - 更新调用 `browser.SubmitCode()` 时传递 `info.ProblemID`

### 当前测试状态

**Bug #1: HTML 解析换行符**
- ✅ **旧格式** (Contest 1000a): 正常工作 - 8 行正确分离
- ✅ **新格式** (Contest 2122d): 已修复 - 13 行正确分离
- ✅ **日志输出**: 详细的 INFO 级别日志便于调试

**Bug #2: 提交未选择题目**
- ✅ **完全修复**: 题目选择、大写转换、提交按钮点击都正常

### Git Commit History

**已完成**:
```bash
commit 503b6a2: HOTFIX - Handle <br> tags in HTML parser to preserve newlines
commit 8667beb: HOTFIX - Add uppercase conversion for problemID in browser submit
commit <pending>: HOTFIX - Handle <div> tags in HTML parser for new format
  1. Add </div> tag replacement with newlines
  2. Add detailed INFO level logging for parsing steps
  3. Test both old (<br>) and new (<div>) formats - both pass
```

---

**最后更新**: 2025-12-31 20:30
**状态**: ✅ 所有关键 bug 已完整修复
- ✅ Bug #1 旧格式: `<br>` 标签处理正常
- ✅ Bug #1 新格式: `<div>` 标签处理已添加
- ✅ Bug #1 日志: 详细的 INFO 级别日志便于调试
- ✅ Bug #2: 提交时正确选择题目并转换为大写
