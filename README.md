# CodeForge

Enterprise coding Agent CLI — **test/quality engineering wedge**（企业测试/QE CLI 楔子，**不是**通用 IDE 或聊天框）。

M1 在 W1 骨架上把三条垂直命令做成真实路径（黄金路径 + 失败路径），并加上本地席位状态机。

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

适配器走 Chat Completions：`POST {base_url}/chat/completions`。默认值仍是兼容占位；**请按网关配置** `base_url` / `model`。

DeepSeek 示例：

```bash
export CODEFORGE_API_KEY=sk-...
export CODEFORGE_BASE_URL=https://api.deepseek.com/v1
export CODEFORGE_MODEL=deepseek-chat
./codeforge provider ping
```

其他兼容网关同样设置 `CODEFORGE_BASE_URL` + `CODEFORGE_MODEL`（或写入 `~/.codeforge/config.yaml`）。

无密钥时，三条垂直命令走 **本地确定性分析**（CI 友好）。有密钥且未加 `--offline` 时，会额外调用一次 provider；模型编造的文件/函数会被丢弃。

CI 夹具模式：

```bash
export CODEFORGE_OFFLINE=1
# or
./codeforge test-gen --offline --module samples/go-saas-admin
```

## Seat lifecycle（席位：trial → active → suspended）

本地账本：`~/.codeforge/seat.json`（不是云计费）。**不新增第六条产品命令**，挂在 `login` 子命令上。

```bash
./codeforge login                 # 无席位则进入 trial，并打印状态
./codeforge login status          # 只读
./codeforge login trial           # (none) → trial
./codeforge login activate        # trial → active
./codeforge login suspend         # active → suspended
```

非法跳转（例如 trial→suspend、suspended→active）非 0 退出且不改账本。

`suspended` 时 `test-gen` / `defect-blame` / `regress-suggest` 拒绝执行。首次跑垂直命令若还没有席位，会自动开 trial。

## Commands

Binary: `codeforge`. 产品命令仍是这五条（加深行为，不是聊天 IDE）：

| Command | M1 |
|---|---|
| `login` | 席位账本 + 凭据提示 |
| `init` | 打印配置路径（不写文件） |
| `test-gen` | 扫描模块，产出有证据的测例 |
| `defect-blame` | 堆栈归因 + 复现步骤 |
| `regress-suggest` | P0/P1/P2 定向回归（禁止全量） |

诊断（非产品命令）：`provider ping`。

### test-gen

```bash
./codeforge test-gen --offline --module samples/go-saas-admin
./codeforge test-gen --offline --module samples/java-saas-admin
```

- **黄金路径**：对样例订单域扫描 `Refund`/`Get`，给出关闭订单退款拒绝 + 租户隔离用例（CF-W1-001）。
- **失败路径**：空模块（无生产源码）**非 0 退出**，不编造测例。

### defect-blame

```bash
./codeforge defect-blame --offline --stack fixtures/failure-stack-zh.txt
./codeforge defect-blame --offline --stack fixtures/insufficient-stack.txt
```

- **黄金路径**：定位 `OrderService.refund`，分类 `defect`，给出租户/订单复现步骤。
- **失败路径**：无应用帧 / 空堆栈 → `insufficient`，**不编造**文件或责任人，非 0 退出。

### regress-suggest

```bash
./codeforge regress-suggest --offline --diff fixtures/fake-pr.diff
./codeforge regress-suggest --offline --diff fixtures/docs-only.diff
```

- **黄金路径**：P0 退款冒烟、P1 租户核心、可选 P2；输出含 “full regression forbidden”。
- **失败路径**：空 diff 非 0 退出；**纯文档**不升级为 P0/P1。

## Config

合并顺序（后者覆盖前者）：

1. Defaults（`base_url=https://api.openai.com/v1`，`model=gpt-4o-mini` — 可被网关覆盖）
2. `~/.codeforge/config.yaml`
3. 项目 `.codeforge.yaml`（从 cwd 向上查找）
4. Env：`CODEFORGE_API_KEY`，`CODEFORGE_BASE_URL`，`CODEFORGE_MODEL`

```yaml
# ~/.codeforge/config.yaml — DeepSeek 兼容网关示例
api_key: sk-...
base_url: https://api.deepseek.com/v1
model: deepseek-chat
```

不要把密钥提交进仓库。

## Samples and fixtures

- `samples/java-saas-admin/` / `samples/go-saas-admin/` — SaaS 中后台订单骨架（已知缺陷 CF-W1-001）
- `fixtures/failure-stack-zh.txt` — 中文业务失败堆栈
- `fixtures/insufficient-stack.txt` — 证据不足（超时、无应用帧）
- `fixtures/fake-pr.diff` — 假 PR：退款账本但漏状态/租户
- `fixtures/docs-only.diff` — 纯文档变更（不得升级回归）

## Scope

CodeForge **只做**测例生成 / 缺陷归因 / 回归建议。做成通用 IDE、编码 Copilot 或聊天框视为未达立项。无 IDE 插件。

选型：[SELECTION.md](SELECTION.md)。
