# CodeForge

Enterprise coding Agent CLI — **test/quality engineering wedge** (not a general IDE or chat box).

W1 delivers an installable CLI skeleton, a frozen provider/config/sample layout, and demo fixtures. Test generation, defect attribution, and regression suggestions are stubs until W2.

## Build / Install

Requires Go 1.22+.

```bash
git clone https://github.com/mxl1c/CodeForge.git
cd CodeForge
go build -o codeforge ./cmd/codeforge
```

Install into `GOBIN` / `PATH`:

```bash
go install github.com/mxl1c/CodeForge/cmd/codeforge@latest
```

From a local checkout:

```bash
go install ./cmd/codeforge
```

## Commands

Binary name: `codeforge`. W1 commands are stubs: they print a clear status line and **exit 0**.

| Command | Purpose |
|---|---|
| `codeforge login` | Auth stub (no-op). Set `CODEFORGE_API_KEY` (optional `CODEFORGE_BASE_URL`). |
| `codeforge init` | Project config stub (does not write files in W1). |
| `codeforge test-gen` | Test generation stub. |
| `codeforge defect-blame` | Defect attribution stub. |
| `codeforge regress-suggest` | Regression-suggestion stub. |

```bash
./codeforge --help
./codeforge login
./codeforge init
./codeforge test-gen
./codeforge defect-blame
./codeforge regress-suggest
```

`test-gen` / `defect-blame` / `regress-suggest` construct the OpenAI-compatible provider. If no API key is configured they print a graceful error and still exit 0 (stub). They do not call the model in W1.

## Config

Merged in this order (later wins):

1. Defaults (`base_url=https://api.openai.com/v1`, `model=gpt-4o-mini`)
2. `~/.codeforge/config.yaml`
3. Project `.codeforge.yaml` (walks up from the current directory)
4. Env: `CODEFORGE_API_KEY`, `CODEFORGE_BASE_URL`

Example user config (`~/.codeforge/config.yaml`):

```yaml
api_key: sk-...
base_url: https://api.openai.com/v1
model: gpt-4o-mini
```

Do not commit secrets. Project `.codeforge.yaml` should only hold non-secret defaults.

## Samples and fixtures

- `samples/java-saas-admin/` — Java SaaS mid-office stub (order admin) with a failing-test hook
- `samples/go-saas-admin/` — Go SaaS mid-office stub with the same known defects
- `fixtures/failure-stack-zh.txt` — Chinese business-context failure stack
- `fixtures/fake-pr.diff` — fake PR that introduces the refund regression

Known defect `CF-W1-001`: closed orders can be refunded; order lookup is not tenant-scoped.

## Scope

CodeForge is a **test/quality wedge**: test generation, defect blame, and regression suggestions for Java/Go SaaS mid-office services. It is **not** a general-purpose IDE, coding copilot, or chat interface.

Frozen choices: [SELECTION.md](SELECTION.md).
