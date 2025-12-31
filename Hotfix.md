# CF-Tool Hotfix 记录

> **创建日期**: 2025-12-31
> **版本**: v1.0-browser
> **状态**: 进行中

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

**文件**: `client/html/parser.go:40-66`

**问题 1**: `<br>` 标签被直接删除（主要问题）

Codeforces 的 HTML 使用 `<br />` 标签来表示换行：
```html
<pre>3<br />XS<br />XS<br />M<br />XL<br />S<br />XS<br /></pre>
```

原代码直接删除所有 HTML 标签（包括 `<br />`），导致换行符丢失：
```go
// Remove all HTML tags
tagReg := regexp.MustCompile(`<[^>]+>`)
text = tagReg.ReplaceAllString(string(htmlBytes), "")  // ❌ 删除了 <br /> 标签
```

**问题 2**: `\s+` 匹配并替换换行符（次要问题，已在第一版修复）

原代码使用 `\s+` 将所有空白符（包括 `\n`）替换成空格：
```go
spaceReg := regexp.MustCompile(`\s+`)
text = spaceReg.ReplaceAllString(text, " ")  // ❌ 把 \n 也替换成空格
```

**示例**:

HTML 输入：
```html
<pre>3<br />XS<br />XS<br />M<br /></pre>
```

**原代码处理（错误）**:
```
1. 删除所有标签: "3XSXSM"
2. HTML unescape: "3XSXSM"
3. \s+ 替换空白符: "3XSXSM"
Result: "3XSXSM"  # ❌ 所有内容在一行
```

**修复后处理（正确）**:
```
1. 替换 <br> 为 \n: "3\nXS\nXS\nM\n"
2. 删除其他标签: "3\nXS\nXS\nM\n"
3. HTML unescape: "3\nXS\nXS\nM\n"
4. [ \t]+ 只替换空格/制表符: "3\nXS\nXS\nM\n"
5. Trim 每行: "3\nXS\nXS\nM\n"
Result: "3\nXS\nXS\nM\n"  # ✅ 换行符正确保留
```

### 修复方案

**关键修复**: 在删除其他 HTML 标签之前，先将 `<br>` 标签替换为换行符 `\n`

```go
func extractTextContent(htmlBytes []byte) string {
    text := string(htmlBytes)

    // CRITICAL: Replace <br> tags with newlines BEFORE removing other tags
    // This handles both <br>, <br/>, and <br /> variants
    brReg := regexp.MustCompile(`<br\s*/?>`)
    text = brReg.ReplaceAllString(text, "\n")

    // Remove all remaining HTML tags
    tagReg := regexp.MustCompile(`<[^>]+>`)
    text = tagReg.ReplaceAllString(text, "")

    // Unescape HTML entities
    text = html.UnescapeString(text)

    // ONLY replace spaces and tabs, NOT newlines or carriage returns
    spaceReg := regexp.MustCompile(`[ \t]+`)
    text = spaceReg.ReplaceAllString(text, " ")

    // Trim each line and preserve line breaks
    lines := strings.Split(text, "\n")
    for i, line := range lines {
        lines[i] = strings.TrimSpace(line)
    }
    text = strings.Join(lines, "\n")

    // Trim leading/trailing whitespace (spaces/tabs) but keep structure
    text = strings.Trim(text, " \t\r")

    return text
}
```

#### 方案 2: 使用 HTML 规范化（更健壮）

```go
func extractTextContent(htmlBytes []byte) string {
    // Use goquery for better HTML parsing
    // ...
}
```

**选择**: 方案 1（更简单，无需额外依赖）

### 测试计划

1. 测试多行输入
2. 测试包含空行的输入
3. 测试只有一行的输入
4. 测试输出（通常也是多行）

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

### Bug #1: 换行符丢失
- [x] 问题调查
- [x] 修改 `client/html/parser.go` (2025-12-31 19:44)
- [x] 测试多行输入 (2025-12-31 19:44)
- [x] 验证输出文件 (2025-12-31 19:44)

**测试结果**:
```bash
$ cf parse 1000 a

# 修复前（错误）
in1.txt: "3XSXSMXLSXS"  # 所有内容在一行

# 修复后（正确）
in1.txt:
3
XS
XS
M
XL
S
XS

# 验证字节内容
$ od -c cf/contest/1000/a/in1.txt
0000000   3  \n   X   S  \n   X   S  \n   M  \n   X   L  \n   S  \n   X
0000020   S  \n  \n
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

## ✅ 修复总结

### 修改的文件

1. **client/html/parser.go**
   - 修改 `extractTextContent()` 函数
   - **关键修复**: 添加 `<br>` 标签处理，在删除其他标签前先替换为 `\n`
   - 将 `\s+` 改为 `[ \t]+`，只替换空格和制表符，保留换行符
   - 添加逐行 trim 逻辑

2. **client/browser/submit.go**
   - 添加 `problemID` 参数到 `SubmitCode()` 函数
   - 添加选择题目的步骤（Step 2）
   - **添加大写转换**: `problemIDUpper := strings.ToUpper(problemID)` 确保符合 Codeforces 要求
   - 更新提交按钮选择器，添加 `#singlePageSubmitButton`
   - 重新组织步骤顺序：选择题目 → 选择语言 → 注入代码 → 点击提交

3. **client/submit.go**
   - 更新调用 `browser.SubmitCode()` 时传递 `info.ProblemID`

### 测试验证

✅ **Bug #1 修复验证**:
- 测试单行输入：正常
- 测试多行输入：换行符正确保留
- 测试输出文件：格式正确

✅ **Bug #2 修复验证**:
- 提交 A 题：成功
- **大写转换**: "a" → "A" 正确转换
- 选择题目正确填充：`submittedProblemIndex` = "A"
- 提交按钮点击成功：使用 `#singlePageSubmitButton`
- 提交记录显示正确的题目：`problem=A - Homework`

### Git Commit

```bash
commit: HOTFIX - Fix critical bugs in parse and submit
1. HTML parser: Handle <br> tags to preserve newlines (Bug #1)
2. Browser submit: Select problem before submitting (Bug #2)
3. Browser submit: Convert problemID to uppercase (a → A)
```

**关键发现**: Bug #1 的真正原因是 HTML 中使用 `<br />` 标签表示换行，而不是 `\n` 字符。
必须在删除其他 HTML 标签之前先将 `<br>` 替换为换行符。

---

**最后更新**: 2025-12-31 20:15
**状态**: ✅ 所有关键 bug 已修复并测试通过
- Bug #1: HTML parser 正确处理 `<br>` 标签保留换行符
- Bug #2: 提交时正确选择题目并转换为大写
