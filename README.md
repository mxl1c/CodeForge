# CodeForge

Enterprise coding **agent CLI** for the **test / QA** wedge — not a general chat app, not an IDE.

W1 ships a runnable skeleton: Cobra commands, a switchable model provider (OpenAI-compatible), dual config files, Java + Go SaaS-admin samples, and Chinese business-context fixtures.

## Frozen W1 selection

| Layer | Choice |
| --- | --- |
| Language / runtime | **Go 1.22+** |
| CLI | **Cobra** (`codeforge`) |
| Model provider | **Interface + OpenAI-compatible adapter** |
| Credentials | env `CODEFORGE_API_KEY` / `CODEFORGE_BASE_URL` (optional `CODEFORGE_MODEL`) |
| User config | `~/.codeforge/config.yaml` |
| Project config | `.codeforge.yaml` |
| Distribution this week | **CLI only** — no IDE plugins |
| Product wedge | test generation, defect blame, regression suggestions |

See [docs/w1-selection.md](docs/w1-selection.md) for the full lock.

## Install

```bash
go 1.22+ required

git clone https://github.com/mxl1c/CodeForge.git
cd CodeForge
go build -o bin/codeforge ./cmd/codeforge
```

## CLI (five subcommands)

```text
codeforge --help
codeforge login            # write ~/.codeforge/config.yaml
codeforge init             # write project .codeforge.yaml
codeforge test-gen         # test generation skeleton (+ optional live model call)
codeforge defect-blame     # failure-stack → likely owners (skeleton)
codeforge regress-suggest  # PR diff → regression cases (skeleton)
```

Stubs print usage-quality placeholders. They are intentionally not a full agent loop.

### Examples against shipped fixtures

```bash
./bin/codeforge test-gen --path samples/java-saas-admin
./bin/codeforge defect-blame --stack fixtures/failure-stack-zh.txt
./bin/codeforge regress-suggest --diff fixtures/fake-pr.diff
```

## Provider (switchable)

Commands talk to `internal/provider.Provider`, not to a vendor SDK.

W1 registers one adapter:

- **`openai-compatible`** — HTTP `POST {baseURL}/chat/completions`

Swap later by implementing `Provider` and adding a branch in `internal/provider.New`.

Factory:

```go
p, err := provider.New(provider.Config{
    Provider: "openai-compatible", // default
    APIKey:   os.Getenv("CODEFORGE_API_KEY"),
    BaseURL:  os.Getenv("CODEFORGE_BASE_URL"),
    Model:    "gpt-4o-mini",
})
```

## Model API: one real call path

The live HTTP path is `OpenAI.Chat` → `{baseURL}/chat/completions`.

`test-gen --live` is the command that exercises it.

**Skip live call when no key (default):**

```bash
./bin/codeforge test-gen --path samples/go-saas-admin
# prints: skipping live model call: no API key
```

**Perform a live call:**

```bash
export CODEFORGE_API_KEY=sk-...
export CODEFORGE_BASE_URL=https://api.openai.com/v1   # any OpenAI-compatible host
export CODEFORGE_MODEL=gpt-4o-mini                    # optional
./bin/codeforge test-gen --path samples/java-saas-admin --live
```

Or persist credentials (file mode `0600`):

```bash
./bin/codeforge login --api-key sk-... --base-url https://api.openai.com/v1
./bin/codeforge test-gen --path samples/java-saas-admin --live
```

Precedence: **environment > `~/.codeforge/config.yaml` > `.codeforge.yaml` > defaults**.

The HTTP adapter is covered by `internal/provider` tests using `httptest` so CI does not need a real key.

## Config

**User** (`~/.codeforge/config.yaml`), created by `codeforge login`:

```yaml
provider: openai-compatible
api_key: sk-...
base_url: https://api.openai.com/v1
model: gpt-4o-mini
```

**Project** (`.codeforge.yaml`), created by `codeforge init`:

```yaml
project: java-saas-admin
language: java
test:
  framework: junit5
```

## Samples and fixtures

| Path | Role |
| --- | --- |
| `samples/java-saas-admin/` | Java SaaS admin slice (order approval, discount, tenant rules) |
| `samples/go-saas-admin/` | Go mirror of the same slice |
| `fixtures/failure-stack-zh.txt` | Chinese business-context failure stack |
| `fixtures/fake-pr.diff` | Fake PR unified diff for regression suggestions |

## Layout

```text
cmd/codeforge/          # main
internal/cli/           # cobra commands
internal/config/        # user + project YAML
internal/provider/      # Provider interface + OpenAI-compatible adapter
samples/java-saas-admin/
samples/go-saas-admin/
fixtures/
docs/w1-selection.md
```

## Test

```bash
go test ./...
go test ./samples/go-saas-admin/...
```

## Out of scope (W1)

- IDE / editor plugins
- General-purpose chat
- Full agent planning / tool loop
- Shipping generated tests back into the samples automatically
