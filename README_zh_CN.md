# Codeforces Tool

[![Build Status](https://travis-ci.org/xalanq/cf-tool.svg?branch=master)](https://travis-ci.org/xalanq/cf-tool)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/cf-tool)](https://goreportcard.com/report/github.com/xalanq/cf-tool)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.12-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/cf-tool/master/LICENSE)

Codeforces Tool 是 [Codeforces](https://codeforces.com) 的命令行界面的工具。

这是原版 [cf-tool](https://github.com/xalanq/cf-tool) 的一个分支，增加了浏览器模式支持以绕过 Cloudflare 保护。

[English](./README.md)

## 安装

### 第一步：编译 cf-tool

克隆仓库到本地，然后使用 make 编译 **(go >= 1.12)**：

```bash
# 克隆仓库
git clone https://github.com/NetWilliam/cf-tool.git
cd cf-tool

# 使用 make 编译
make build

# （可选）安装到 ~/go/bin
make install
```

编译后的二进制文件位于 `./bin/cf`。你可以把它移动到任何位置，或者把它添加到系统 PATH 环境变量中。

### 第二步：安装浏览器模式组件

⚠️ **重要提示**：浏览器模式是所有网络操作（parse、submit、list、watch、race、pull、clone）的**必需要求**。

cf-tool 使用**浏览器模式**来绕过 Codeforces 的 Cloudflare 保护。你需要安装：

1. **[mcp-chrome](https://github.com/hangwin/mcp-chrome/)** - 通过 MCP 暴露 Chrome DevTools Protocol 的 Chrome 扩展

2. **mcp-chrome-bridge** - Node.js 桥接服务

#### 快速安装

```bash
# 安装 mcp-chrome-bridge
npm install -g @hangwin/mcp-chrome-bridge

# 或使用 pnpm
pnpm add -g @hangwin/mcp-chrome-bridge
```

然后：

1. 从 [GitHub releases](https://github.com/hangwin/mcp-chrome/releases) 下载 mcp-chrome 扩展
2. 解压到文件夹
3. 打开 Chrome 浏览器，访问 `chrome://extensions/`
4. 开启"开发者模式"
5. 点击"加载已解压的扩展程序"，选择扩展文件夹
6. 在终端运行 `mcp-chrome-bridge`（运行在 `http://127.0.0.1:12306/mcp`）

#### 验证安装

```bash
# 测试 MCP 连接
cf mcp-ping

# 测试浏览器自动化
cf mocka
```

**重要提示**：在使用 cf-tool 之前，请确保两个命令都成功运行！

更多关于 mcp-chrome 的详情，请访问：https://github.com/hangwin/mcp-chrome/

## 新增命令

这个分支新增了两个用于测试浏览器模式的命令：

### cf mcp-ping

测试与 MCP Chrome Server 的连接。

```bash
cf mcp-ping
```

预期输出：
```
✅ MCP Chrome Server is running
```

### cf mocka

测试浏览器自动化功能。打开 Chrome 浏览器，导航到 Google 并搜索 "billboard 季度榜"，然后返回页面内容以验证浏览器自动化是否正常工作。

```bash
cf mocka
```

此命令用于验证 cf-tool 能够正确控制你的浏览器。

## 已验证命令

以下命令已测试并验证在浏览器模式下正常工作：

- [x] `cf parse` - 获取题目样例
- [x] `cf gen` - 从模板生成代码
- [x] `cf test` - 本地编译和测试
- [x] `cf submit` - 提交代码到 Codeforces
- [x] `cf open` - 在浏览器中打开题目
- [x] `cf sid` - 打开提交页面
- [x] `cf race` - 比赛倒计时和解析
- [ ] `cf clone` - 克隆用户提交（~~部分工作~~，行为不明确）

## 原版文档

关于所有 cf-tool 命令的详细用法（parse、submit、race 等），请参考[原版仓库](https://github.com/xalanq/cf-tool)。

原版的所有命令在此分支中都得到支持，并且网络操作默认启用浏览器模式。

## 常见问题

### 浏览器模式是必需的吗？

是的。由于 Codeforces 的 Cloudflare 保护，所有依赖网络的命令现在都需要浏览器模式。

### 如何检查浏览器模式是否正常工作？

运行 `cf mcp-ping`。如果显示 "✅ MCP Chrome Server is running"，说明浏览器模式已就绪。

### 我可以使用旧的 HTTP 模式吗？

不可以。旧的 HTTP 模式无法绕过 Cloudflare 保护，不再被支持。

## 许可证

MIT License - 与原版 [cf-tool](https://github.com/xalanq/cf-tool) 相同
