# W1 selection (frozen)

CodeForge W1 is a **test/QA coding-agent CLI**, not a general assistant and not an IDE.

## Product

- Vertical: enterprise **test generation**, **defect blame**, **regression suggestion**.
- Surface this week: **CLI only**.
- Languages in samples: **Java** and **Go** SaaS-admin slices.
- Fixtures include a **Chinese business-context** failure stack.

## Tech lock

| Decision | Value | Why it is locked |
| --- | --- | --- |
| Language | Go 1.22+ | Single static binary, matches the Go sample, fast CI |
| CLI framework | Cobra | Subcommands `login`, `init`, `test-gen`, `defect-blame`, `regress-suggest` |
| Provider | `Provider` interface + OpenAI-compatible HTTP adapter | Vendor-switchable without rewriting commands |
| Auth env | `CODEFORGE_API_KEY`, `CODEFORGE_BASE_URL` | Works with OpenAI and compatible gateways |
| User config | `~/.codeforge/config.yaml` | Credentials stay out of the repo |
| Project config | `.codeforge.yaml` | Language / test framework per sample |
| IDE plugins | **Not shipped** | W1 gate is CLI-only |

## Commands (exact names)

1. `login` — store credentials in the user config file.
2. `init` — write project `.codeforge.yaml`.
3. `test-gen` — test-generation skeleton; optional `--live` model call.
4. `defect-blame` — stack → owner placeholder using `fixtures/failure-stack-zh.txt`.
5. `regress-suggest` — diff → regression cases using `fixtures/fake-pr.diff`.

## Acceptance mapping

| Gate | Where it lives |
| --- | --- |
| Selection reflected in README / docs | `README.md`, this file |
| Five subcommand skeletons run | `internal/cli/*`, `cmd/codeforge` |
| Provider switchable | `internal/provider.Provider` + `New` |
| Both sample paths + both fixtures | `samples/*`, `fixtures/*` |
| Real model call path wired once | `OpenAI.Chat`; `test-gen --live`; skip if no key |
