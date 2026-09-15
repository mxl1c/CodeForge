# CodeForge W1 技术选型冻结

本文锁定 Week 1 骨架的技术与产品边界。后续周次不得在未更新本文的情况下突破冻结项。

## 产品楔子（Wedge）

| 项 | 冻结 |
| --- | --- |
| 市场 | 中国企业中后台（SaaS Admin / 租户 / 审批 / 工单） |
| 交付 | 席位制（seat），按使用 Agent 的 QE/研发席位计费与授权 |
| 垂直 | 测试与质量工程（test / QE）：测试生成、缺陷归因、回归建议 |
| 非目标 | 通用聊天 App；Cursor 克隆 IDE；本周不做 IDE 插件 |

## 运行时与语言

- **Go 1.22+**（本仓库 `go.mod` 的 `go` 指令为 1.22）
- 仅 CLI；入口：`cmd/codeforge`
- 构建：`go build -o codeforge ./cmd/codeforge`
- 测试：`go test ./...`

## CLI

- 框架：**cobra**
- 五个产品命令（必须出现在 `--help` 中）：
  1. `codeforge login` — 写入席位凭据到用户配置
  2. `codeforge init` — 生成项目 `.codeforge.yaml`
  3. `codeforge test-gen` — 测试生成（W1 stub）
  4. `codeforge defect-blame` — 缺陷归因（W1 stub）
  5. `codeforge regress-suggest` — 回归建议（W1 stub）
- Provider 探测（等价命令，W1 必达）：`codeforge provider ping`

W1 要求：`--help` 与上述 stub 命令干净退出（无 panic）。`provider ping` 在缺少 Key 时返回明确错误；Key 已配置时发起一次真实 HTTP 调用。

## Provider

- 抽象：`internal/provider.Provider`（`Ping` / `Complete`）
- 适配器：OpenAI 兼容 Chat Completions（`/chat/completions`）
- 凭据：
  - 必选：`CODEFORGE_API_KEY`（优先于用户配置文件）
  - 可选：`CODEFORGE_BASE_URL`（兼容网关 / 私有化推理；默认 `https://api.openai.com/v1`）
- 缺 Key：返回 `provider.ErrMissingAPIKey`，文案说明 env、`login`、以及 `~/.codeforge/config.yaml`

## 配置路径

| 层级 | 路径 | 内容 |
| --- | --- | --- |
| 用户 | `~/.codeforge/config.yaml` | `api_key` / `base_url` / `model`（权限 0600） |
| 项目 | `.codeforge.yaml` | 项目名、语言、楔子、样例与 fixture 路径 |
| 测试覆盖 | `$CODEFORGE_HOME` | 覆盖用户配置目录，避免测试写入真实 `$HOME` |

合并优先级：环境变量 > 用户配置文件 > 内置默认值。

## 样例与 Fixture 路径（锁定）

不得改名或挪路径：

- `samples/java-saas-admin/` — 脱敏 Java SaaS 中后台骨架
- `samples/go-saas-admin/` — 脱敏 Go SaaS 中后台骨架
- `fixtures/failure-stack-zh.txt` — 中文业务失败栈
- `fixtures/fake-pr.diff` — 伪 PR diff（含状态机回归点）

## 本周明确不做

- IDE / VS Code / JetBrains 插件
- 通用对话产品形态
- 非 QE 的“写业务代码”主路径
- 真实客户代码或未脱敏数据
