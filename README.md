# CodeForge

企业测试 / 质量工程（QE）Agent **CLI**。面向中国企业中后台，按**席位**交付。  
不是通用聊天 App，也不是 Cursor 克隆 IDE。W1 仅 CLI。

Enterprise coding Agent CLI for the **test/QE** wedge (Chinese enterprise SaaS admin). Seat-based. CLI-only this week.

## 安装 / Install

要求 Go 1.22+。

```bash
git clone https://github.com/mxl1c/CodeForge.git
cd CodeForge
go build -o codeforge ./cmd/codeforge
./codeforge --help
```

或：

```bash
go install github.com/mxl1c/CodeForge/cmd/codeforge@latest
```

## API Key

使用 OpenAI 兼容接口。Key 缺失时 `codeforge provider ping` 会给出明确错误。

```bash
export CODEFORGE_API_KEY="sk-..."
# 可选：私有化 / 兼容网关
export CODEFORGE_BASE_URL="https://api.openai.com/v1"

# 或写入用户配置 ~/.codeforge/config.yaml
./codeforge login --api-key "sk-..."
```

探测（有 Key 时发起一次真实调用）：

```bash
./codeforge provider ping
```

项目配置（当前目录 `.codeforge.yaml`）：

```bash
./codeforge init
```

## 五个命令 / Five commands

| 命令 | 作用 |
| --- | --- |
| `codeforge login` | 将席位凭据写入 `~/.codeforge/config.yaml` |
| `codeforge init` | 生成项目 `.codeforge.yaml` |
| `codeforge test-gen` | 为中后台代码生成测试（W1 stub） |
| `codeforge defect-blame` | 根据失败栈做缺陷归因（W1 stub） |
| `codeforge regress-suggest` | 根据 PR diff 建议回归范围（W1 stub） |

```bash
./codeforge login --help
./codeforge init --help
./codeforge test-gen --help
./codeforge defect-blame --help
./codeforge regress-suggest --help
```

W1 stub 可带样例路径演练：

```bash
./codeforge test-gen --path samples/java-saas-admin --lang java
./codeforge defect-blame --stack fixtures/failure-stack-zh.txt
./codeforge regress-suggest --diff fixtures/fake-pr.diff
```

## 仓库布局

- `docs/SELECTION.md` — W1 技术选型冻结
- `samples/java-saas-admin/`、`samples/go-saas-admin/` — 脱敏中后台骨架
- `fixtures/failure-stack-zh.txt`、`fixtures/fake-pr.diff` — QE 演练夹具
- `internal/provider` — Provider 接口 + OpenAI 兼容适配器

## 开发

```bash
go test ./...
go build ./...
```
