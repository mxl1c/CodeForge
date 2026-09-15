# CodeForge 选型冻结（W1）

> 状态：已冻结，W1 骨架按此实现。变更须走显式解冻，不得在实现中顺手改。

## 产品楔子

- **只做**测试垂类闭环 + CLI 接入（测例生成 / 缺陷归因 / 回归建议）。
- **不是**通用 IDE、编码 Copilot 或聊天框。做成通用 IDE 视为未达立项。
- 首发 **CLI**；VS Code / JetBrains 插件为 P1，不进入 M0–M1 主盘。

## 语言与 CLI

| 项 | 冻结值 |
|---|---|
| 语言 | Go **1.22+** |
| CLI 框架 | [spf13/cobra](https://github.com/spf13/cobra) |
| 二进制名 | `codeforge` |
| 产品子命令（仅此五条） | `login` · `init` · `test-gen` · `defect-blame` · `regress-suggest` |
| W1 产品命令行为 | 五条均为 stub：明确输出 + **exit 0**；**不**调用模型 |
| 诊断命令（非产品） | `provider ping`（W1-05 API smoke） |

不增加第六条产品命令（含 `version` 作为独立产品命令）。Cobra 自带 `help` 保留。`provider ping` 是诊断/冒烟入口，不计入产品命令范围。

## Provider

| 项 | 冻结值 |
|---|---|
| 抽象 | `internal/provider.Provider`（`Complete`） |
| 适配器 | OpenAI-compatible Chat Completions（`POST {base_url}/chat/completions`） |
| 鉴权 | `CODEFORGE_API_KEY`；缺失时返回 `provider.ErrNoAPIKey`（不 panic） |
| Base URL | `CODEFORGE_BASE_URL`，默认 `https://api.openai.com/v1` |
| 默认模型名 | `gpt-4o-mini`（可被配置覆盖；实际模型由兼容网关决定） |

W1 接通适配器与错误路径；不在 stub 命令里真实跑测例生成。

W1-05：`codeforge provider ping` 在凭据存在时发起 **一次** 真实 Chat Completions `Complete`（`POST {base_url}/chat/completions`），成功则打印 model / token usage 并以 0 退出；缺失 `CODEFORGE_API_KEY`（及配置 `api_key`）时明确报错、非 0 退出。产品五命令仍不得调用模型。

## 配置

| 项 | 冻结值 |
|---|---|
| 用户配置 | `~/.codeforge/config.yaml` |
| 项目配置 | `.codeforge.yaml`（从 cwd 向上查找） |
| 合并顺序 | 默认值 &lt; 用户配置 &lt; 项目配置 &lt; 环境变量 |
| 密钥 | 禁止写入仓库；项目文件只放非密钥字段 |

字段（最小集）：`api_key`、`base_url`、`model`。

## 样例仓与夹具

| 路径 | 冻结用途 |
|---|---|
| `samples/java-saas-admin/` | Java SaaS 中后台骨架（租户订单），含失败测钩子 |
| `samples/go-saas-admin/` | Go SaaS 中后台骨架，缺陷与 Java 样例对齐 |
| `fixtures/failure-stack-zh.txt` | 中文业务上下文失败堆栈（订单退款 / 租户） |
| `fixtures/fake-pr.diff` | 假 PR：接入退款账本但漏状态校验与租户隔离 |

样例必须脱敏（无真实客户数据、无密钥），可公开给试点演示。

已知缺陷编号：**CF-W1-001**（关闭订单仍可退款 + 跨租户读取）。

## 明确不做（W1 / M0）

- 第二种接入（IDE 插件）
- 通用聊天 / Agent 工作区
- 云市场 / 在线支付（M2 前计费 = 自建账本 + 线下合同，W1 不实现计费）
- 冻结 JSON Schema 的完整测例 I/O（W2）
- 真实鉴权与写入配置文件（`login` / `init` 在 W1 为 stub）

## 计量字段（选型占位，W1 不实现）

后续自建账本预留：租户、席位、命令名、provider、model、prompt/completion tokens、请求状态。W1 仅在 `CompletionResponse.Usage` 解析 tokens，不上报账本。

---

## M1 解冻（相对 W1 stub）

M1 保持五条产品命令名，加深行为（仍不是通用 IDE / 聊天）：

| 命令 | M1 行为 |
|---|---|
| `login` | 本地席位账本：`trial → active → suspended`（`login trial\|activate\|suspend\|status`） |
| `init` | 打印配置/席位路径，不写入文件 |
| `test-gen --module` | 扫描 Java/Go 生产源码生成测例；空模块失败且不编造 |
| `defect-blame --stack` | 定位文件/函数，分类 defect/env/test；证据不足不编造责任 |
| `regress-suggest --diff` | P0 冒烟 / P1 核心 / P2 外围；禁止全量回归；纯文档不升级 |

有 `CODEFORGE_API_KEY` 时垂直命令可调用一次 provider；`--offline` 或 `CODEFORGE_OFFLINE=1` 为 CI 确定性夹具模式。

Provider 文档优先 **OpenAI 兼容网关**（官方 OpenAI 只是其中一种）。DeepSeek 示例：`CODEFORGE_BASE_URL=https://api.deepseek.com/v1`，`CODEFORGE_MODEL=deepseek-chat`。
