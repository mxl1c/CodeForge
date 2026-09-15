# CodeForge

Enterprise coding Agent CLI — **test/quality engineering wedge**（企业测试/QE CLI 楔子，**不是**通用 IDE 或聊天框）。

M1 锁定垂直命令契约；M2 在席位账本上增加档位字段，并提供自建用量导出（仍是测试/QE 楔子，不是通用 IDE / 聊天 / 在线支付）：

```bash
codeforge test-gen --repo <path>
codeforge defect-blame --log <path>          # default: fixtures/failure-stack-zh.txt
codeforge regress-suggest --diff <path>      # default: fixtures/fake-pr.diff
codeforge seat trial|activate|suspend|status|tier
codeforge usage export --out <file>
```

## Build / Install

Requires Go 1.22+.

```bash
git clone https://github.com/mxl1c/CodeForge.git
cd CodeForge
go build -o codeforge ./cmd/codeforge
```

```bash
go install github.com/mxl1c/CodeForge/cmd/codeforge@latest
# or from a checkout:
go install ./cmd/codeforge
```

## Provider（OpenAI-compatible 网关，不绑死官方 OpenAI）

适配器走 Chat Completions：`POST {base_url}/chat/completions`。**请按网关配置** `base_url` / `model`。官方 OpenAI 只是兼容端点之一。

DeepSeek-compatible default Base URL 示例：

```bash
export CODEFORGE_API_KEY=sk-...
export CODEFORGE_BASE_URL=https://api.deepseek.com/v1
export CODEFORGE_MODEL=deepseek-chat
./codeforge provider ping
```

或写入 `~/.codeforge/config.yaml`：

```yaml
api_key: sk-...
base_url: https://api.deepseek.com/v1
model: deepseek-chat
```

无密钥时，三条垂直命令走 **本地确定性分析**（CI 友好）。有密钥且未加 `--offline` 时，会额外调用一次 provider；模型编造的文件/函数会被丢弃。

```bash
export CODEFORGE_OFFLINE=1
./codeforge test-gen --offline --repo samples/go-saas-admin
```

## Seat lifecycle + tiers

本地账本：`~/.codeforge/seat.json`（**不是**云计费 / **不是**在线支付）。

生命周期仍为 **trial → active → suspended**：

```bash
./codeforge seat trial      # (none) → trial
./codeforge seat activate   # trial → active
./codeforge seat suspend    # active → suspended
./codeforge seat status
```

非法跳转（例如 trial→suspend、suspended→active）非 0 退出且不改账本。`suspended` 时 `test-gen` / `defect-blame` / `regress-suggest` 拒绝执行。首次跑垂直命令若还没有席位，会自动开 trial。

### 档位（offline 标签，不发明支付）

席位账本带商业档位字段，仅用于试点标价展示与 ARPU 分组，**不**走网关、结账或扣款：

| id | 展示 | 含义 |
|---|---|---|
| `free` | Free | 默认 |
| `pro` | ~¥140 | 线下合同档 |
| `business` | ~¥700–1400 | 线下合同档 |

```bash
./codeforge seat tier              # 查看当前档位
./codeforge seat tier free
./codeforge seat tier pro
./codeforge seat tier business
./codeforge seat status            # 同时打印 state 与 tier
```

改档位**不会**改变 trial/active/suspended。

## Usage ledger export（试点 ARPU）

自建用量账本：`~/.codeforge/usage.jsonl`（**不是**云计费）。垂直命令每次调用都会追加一行；`--offline` / stub 路径也会写入**确定性**行（`provider=offline`，`model=offline`，tokens=0），方便演示。有 provider `Complete` 时，prompt/completion tokens 取自响应。

导出字段：`timestamp,tenant,seat_id,tier,command,provider,model,prompt_tokens,completion_tokens,status`。

```bash
./codeforge test-gen --offline --repo samples/go-saas-admin
./codeforge usage export --out arpu.csv
./codeforge usage export --out arpu.json --format json
```

`--format` 可省略：`.json` 导出 JSON，其它扩展名默认 CSV。

`login` 只提示凭据（含 DeepSeek Base URL 示例）；席位走 `seat`，不要用 `login trial`。

## Commands

| Command | Flags / subcommands | 作用 |
|---|---|---|
| `login` | — | 凭据提示（OpenAI 兼容网关） |
| `init` | — | 打印配置路径（不写文件） |
| `test-gen` | `--repo <path>` | 扫描仓库生产源码，产出有证据的测例 |
| `defect-blame` | `--log <path>`（默认 `fixtures/failure-stack-zh.txt`） | 失败日志归因 + 复现步骤 |
| `regress-suggest` | `--diff <path>`（默认 `fixtures/fake-pr.diff`） | P0/P1/P2 定向回归（禁止全量） |
| `seat` | `trial` `activate` `suspend` `status` `tier` | 席位状态机 + 档位（Free / ~¥140 / ~¥700–1400） |
| `usage` | `export --out <file>` | 导出试点 ARPU 用量账本（CSV/JSON） |

诊断：`provider ping`。本产品是测试/QE 楔子，不是通用 IDE 或聊天。

### test-gen

```bash
./codeforge test-gen --offline --repo samples/go-saas-admin
./codeforge test-gen --offline --repo samples/java-saas-admin
```

- **黄金路径**：关闭订单退款拒绝 + 租户隔离用例（CF-W1-001）。
- **失败路径**：空仓库（无生产源码）**非 0 退出**，不编造测例。

### defect-blame

```bash
./codeforge defect-blame --offline
./codeforge defect-blame --offline --log fixtures/failure-stack-zh.txt
./codeforge defect-blame --offline --log fixtures/insufficient-stack.txt
```

- **黄金路径**（默认 log）：定位 `OrderService.refund`，分类 `defect`，给出复现步骤。
- **失败路径**：无应用帧 → `insufficient`，**不编造**责任人，非 0 退出。

### regress-suggest

```bash
./codeforge regress-suggest --offline
./codeforge regress-suggest --offline --diff fixtures/fake-pr.diff
./codeforge regress-suggest --offline --diff fixtures/docs-only.diff
```

- **黄金路径**（默认 diff）：P0 退款冒烟、P1 租户核心；禁止 full regression / 全量回归。
- **失败路径**：空 diff 非 0 退出；**纯文档**不升级为 P0/P1。

## Config

合并顺序（后者覆盖前者）：

1. 代码内占位默认（可被网关覆盖）
2. `~/.codeforge/config.yaml`
3. 项目 `.codeforge.yaml`（从 cwd 向上查找）
4. Env：`CODEFORGE_API_KEY`，`CODEFORGE_BASE_URL`，`CODEFORGE_MODEL`

不要把密钥提交进仓库。

## Samples and fixtures

- `samples/java-saas-admin/` / `samples/go-saas-admin/` — SaaS 中后台订单骨架（CF-W1-001）
- `fixtures/failure-stack-zh.txt` — 默认 defect-blame log
- `fixtures/insufficient-stack.txt` — 证据不足
- `fixtures/fake-pr.diff` — 默认 regress-suggest diff
- `fixtures/docs-only.diff` — 纯文档（不得升级回归）

## Scope

CodeForge **只做**测例生成 / 缺陷归因 / 回归建议 + 席位账本（档位标签）+ 自建用量导出。做成通用 IDE、编码 Copilot、聊天框、在线支付或 Skill 市场视为未达立项。无 IDE 插件。

选型：[SELECTION.md](SELECTION.md)。
