# Plan: gh-ulb — GitHub CLI Extension for User-Level Budgets

## TL;DR

Build `gh-ulb`, a GitHub CLI extension to manage Copilot premium-request User-Level Budgets (ULB) at the enterprise level. Written in **Go** (not bash) for testability, cross-platform support, and first-class `gh` extension conventions. Covers full CRUD on budgets plus batch operations (org team, enterprise team, CSV import). Designed for open-source release with tests, docs, and CI/CD.

## Why Go over Bash

While bash scripts wrapping `gh api` would work, Go is the better choice because:
- **Testing**: Go has `go test` built-in; bash testing (bats/shunit2) is fragile and limited
- **Cross-platform**: Precompiled binaries work on macOS, Linux, Windows without interpreter dependencies
- **Ecosystem**: `github.com/cli/go-gh/v2` provides authenticated API access, config, and output formatting for free
- **Structure**: Cobra CLI framework (used by `gh` itself) gives subcommands, flags, help text, and shell completions
- **Open-source standard**: The vast majority of `gh` extensions (gh-dash, gh-skyline, gh-stack, gh-actions-cache) are Go
- **`gh extension create --precompiled=go`** scaffolds the project with release workflows included

---

## Commands

| Command | Description |
|---|---|
| `gh ulb set-universal` | Create/update the universal (multi_user_customer) budget |
| `gh ulb set-user` | Create a per-user budget override |
| `gh ulb set-team` | Set per-user budgets for all members of an org-level GitHub Team |
| `gh ulb set-enterprise-team` | Set per-user budgets for all members of an Enterprise Team |
| `gh ulb set-csv` | Set per-user budgets from a CSV file |
| `gh ulb list` | List all budgets (optionally filtered to a user) |
| `gh ulb get` | Get the effective budget for a specific user |
| `gh ulb update` | Update an existing budget by ID |
| `gh ulb delete` | Delete a budget by ID |

### Flag Details

**Common flags (all commands):**
- `--enterprise, -e` (required) — enterprise slug
- `--json` — output raw JSON instead of table

**`set-universal`:**
- `--amount, -a` (required) — budget dollar amount (float64)
- hard-stop enforcement is always enabled (`prevent_further_usage: true`)

**`set-user`:**
- `--user, -u` (required) — GitHub username
- `--amount, -a` (required) — budget dollar amount

**`set-team`:**
- `--org, -o` (required) — org name
- `--team, -t` (required) — team slug
- `--amount, -a` (required) — budget amount per user
- `--dry-run` — list users that would be affected without making changes
- `--concurrency` — parallel API calls (default: 5)

**`set-enterprise-team`:**
- `--team, -t` (required) — enterprise team slug
- `--amount, -a` (required) — budget amount per user
- `--dry-run` — preview mode
- `--concurrency` — parallel API calls (default: 5)

**`set-csv`:**
- `--file, -f` (required) — path to CSV file
- `--amount, -a` — default budget amount (used when CSV row has no amount column)
- `--dry-run` — preview mode
- `--concurrency` — parallel API calls (default: 5)

CSV format: `username` column required, optional `amount` column. If `amount` column present, per-row values override the `--amount` flag. Rows without an amount and no `--amount` flag cause an error.

**`list`:**
- `--user` — filter to a specific user
- `--budget-target` — filter by budget target (default: ai_credits, or premium_requests when GH_ULB_USE_PREMIUM_REQUESTS is true)

**`get`:**
- `--user, -u` (required) — GitHub username

**`update`:**
- `--budget-id, -b` (required) — budget ID
- `--amount, -a` (required) — new dollar amount

**`delete`:**
- `--budget-id, -b` (required) — budget ID
- `--confirm` — skip confirmation prompt

---

## Project Structure

```
gh-ulb/
├── main.go                           # Entry point
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go                       # Root cobra command, global flags
│   ├── set_universal.go              # set-universal subcommand
│   ├── set_user.go                   # set-user subcommand
│   ├── set_team.go                   # set-team subcommand
│   ├── set_enterprise_team.go        # set-enterprise-team subcommand
│   ├── set_csv.go                    # set-csv subcommand
│   ├── list.go                       # list subcommand
│   ├── get.go                        # get subcommand
│   ├── update.go                     # update subcommand
│   └── delete.go                     # delete subcommand
├── internal/
│   ├── api/
│   │   ├── client.go                 # HTTP client wrapping go-gh, auth headers
│   │   ├── budget.go                 # Budget CRUD operations
│   │   ├── budget_test.go            # Budget API unit tests (httptest)
│   │   ├── team.go                   # Org team member listing (paginated)
│   │   ├── enterprise_team.go        # Enterprise team member listing (paginated)
│   │   └── team_test.go              # Team resolution tests
│   ├── csv/
│   │   ├── parser.go                 # CSV parsing with validation
│   │   └── parser_test.go
│   ├── batch/
│   │   ├── runner.go                 # Concurrent budget-setting with progress
│   │   └── runner_test.go
│   └── output/
│       ├── formatter.go              # Table/JSON output formatting
│       └── formatter_test.go
├── testdata/
│   ├── valid.csv                     # Test CSV with username,amount
│   ├── username_only.csv             # Test CSV with username only
│   └── invalid.csv                   # Malformed CSV for error tests
├── docs/
│   └── usage.md                      # Extended usage documentation
├── examples/
│   ├── 01-universal-budget.sh        # Set universal budget example
│   ├── 02-individual-override.sh     # Per-user override example
│   ├── 03-team-rollout.sh            # Org team batch example
│   ├── 04-enterprise-team.sh         # Enterprise team batch example
│   ├── 05-csv-import.sh              # CSV import example
│   └── 06-list-and-manage.sh         # List, update, delete examples
├── .github/
│   └── workflows/
│       ├── test.yml                  # CI: go test, go vet, golangci-lint
│       └── release.yml               # Release: gh-extension-precompile action
├── .gitignore
├── .golangci.yml                     # Linter config
├── LICENSE                           # MIT
└── README.md                         # Full documentation
```

---

## Steps

### Phase 1: Scaffold & Core API Client (blocking — everything depends on this)

1. Run `gh extension create --precompiled=go gh-ulb` to bootstrap the project, then replace the generated scaffold with the structure above
2. Set up `go.mod` with dependencies: `github.com/cli/go-gh/v2`, `github.com/spf13/cobra`
3. Implement `cmd/root.go` — root command with `--enterprise` persistent flag, version info
4. Implement `internal/api/client.go` — thin wrapper over `go-gh`'s `gh.RESTClient()` providing authenticated requests with `Accept: application/vnd.github+json` and `X-GitHub-Api-Version: 2022-11-28` headers
5. Implement `internal/api/budget.go` — Budget CRUD functions:
   - `CreateBudget(enterprise, payload) -> Budget` — POST `/enterprises/{enterprise}/settings/billing/budgets`
   - `ListBudgets(enterprise, opts) -> []Budget` — GET same endpoint, with optional `?user=&budgetTarget=` query params
   - `UpdateBudget(enterprise, budgetID, amount) -> Budget` — PATCH `.../{budget_id}`
   - `DeleteBudget(enterprise, budgetID)` — DELETE `.../{budget_id}`
   - Struct types: `Budget`, `BudgetAlerting`, `CreateBudgetParams`
6. Implement `internal/output/formatter.go` — table and JSON rendering for budget data

### Phase 2: CRUD Commands (*depends on Phase 1*)

7. Implement `cmd/set_universal.go` — calls `CreateBudget` with `budget_scope: "multi_user_customer"`
8. Implement `cmd/set_user.go` — calls `CreateBudget` with `budget_scope: "user"` and `user` field
9. Implement `cmd/list.go` — calls `ListBudgets`, renders as table or JSON
10. Implement `cmd/get.go` — calls `ListBudgets` with `user` filter, shows effective budget
11. Implement `cmd/update.go` — calls `UpdateBudget`
12. Implement `cmd/delete.go` — calls `DeleteBudget` with confirmation prompt

### Phase 3: Team Resolution (*parallel with Phase 2*)

13. Implement `internal/api/team.go` — paginated listing of org team members via `GET /orgs/{org}/teams/{team_slug}/members` (per_page=100, follow pagination)
14. Implement `internal/api/enterprise_team.go` — paginated listing of enterprise team members via `GET /enterprises/{enterprise}/teams/{team}/memberships` (per_page=100, follow pagination)
15. Implement `internal/batch/runner.go` — generic concurrent batch processor that takes a list of usernames and an amount, creates user budgets in parallel with configurable concurrency, progress output, error collection, exponential backoff with jitter on 429/5xx (max 3 retries), and idempotent create-or-update logic (POST → 409 → GET → PATCH) with per-user action logging (`[created]`/`[updated]`/`[failed]`) and a final summary

### Phase 4: Batch Commands (*depends on Phase 2 + Phase 3*)

16. Implement `cmd/set_team.go` — resolves org team → usernames via `team.ListMembers()`, then runs batch processor
17. Implement `cmd/set_enterprise_team.go` — resolves enterprise team → usernames, then runs batch processor
18. Implement `internal/csv/parser.go` — parses CSV: validates `username` column exists, optionally reads `amount` column, returns `[]UserBudgetEntry{Username, Amount}`
19. Implement `cmd/set_csv.go` — reads CSV, validates, runs batch processor

### Phase 5: Tests (*parallel with Phase 4, starts during Phase 2*)

20. `internal/api/budget_test.go` — unit tests using `net/http/httptest` to mock API responses for all CRUD operations, error handling (409 for duplicate universal budget), pagination
21. `internal/api/team_test.go` — unit tests for paginated team member listing (both org and enterprise)
22. `internal/csv/parser_test.go` — tests: valid CSV, username-only CSV, missing username column, empty file, malformed rows
23. `internal/batch/runner_test.go` — tests: successful batch, partial failures, dry-run mode, concurrency limits, retry on 429, idempotent update on 409
24. `internal/output/formatter_test.go` — tests: table output, JSON output

### Phase 6: Documentation & CI (*parallel with Phase 4 + Phase 5*)

25. Write `README.md` — installation (`gh extension install`), prerequisites (PAT scopes, enterprise enrollment), all commands with examples, CSV format spec, troubleshooting (common issues from the PDF FAQ)
26. Write `docs/usage.md` — extended guide covering budget hierarchy, how budgets interact with included requests, when budgets reset
27. Write example scripts in `examples/` — runnable shell scripts showing real-world usage patterns
28. Write `.github/workflows/test.yml` — Go CI pipeline: checkout, setup-go, `go vet`, `golangci-lint`, `go test -race -coverprofile=...`
29. Write `.github/workflows/release.yml` — uses `cli/gh-extension-precompile@v2` to auto-build binaries on tag push
30. Add `LICENSE` (MIT), `.gitignore` (binaries, `dist/`), `.golangci.yml`

---

## API Reference (from the ULB Private Preview Guide)

All endpoints are under `https://api.github.com/enterprises/{enterprise}/settings/billing/budgets`

| Method | Path | Purpose |
|---|---|---|
| POST | `/budgets` | Create a budget (universal or user) |
| GET | `/budgets` | List all budgets |
| GET | `/budgets?user={user}&budgetTarget=ai_credits` | Get budgets for a specific user |
| PATCH | `/budgets/{budget_id}` | Update a budget amount |
| DELETE | `/budgets/{budget_id}` | Delete a budget |

**POST body fields:**
- `budget_amount` (float) — dollar amount
- `budget_scope` — `"multi_user_customer"` for universal, `"user"` for individual
- `user` — GitHub username (only when scope is "user")
- `prevent_further_usage` (bool) — block requests when exhausted
- `budget_product_sku` — defaults to `"ai_credits"` (`"premium_requests"` only when `GH_ULB_USE_PREMIUM_REQUESTS` is true)
- `budget_type` — always `"BundlePricing"`
- `budget_alerting` — `{ "will_alert": false, "alert_recipients": [] }` (alerting not yet available)

**Team member APIs:**
- Org teams: `GET /orgs/{org}/teams/{team_slug}/members` — returns `[{login, id, ...}]`, paginated (per_page max 100)
- Enterprise teams: `GET /enterprises/{enterprise}/teams/{team}/memberships` — returns `[{login, id, ...}]`, paginated (per_page max 100)

---

## Key Files

| File | Responsibility |
|---|---|
| `main.go` | Entry point, calls `cmd.Execute()` |
| `cmd/root.go` | Cobra root command with `--enterprise` flag and subcommand registration |
| `internal/api/client.go` | Wraps `go-gh` RESTClient, sets API version header |
| `internal/api/budget.go` | All budget CRUD; struct types `Budget`, `CreateBudgetParams` |
| `internal/api/team.go` | `ListOrgTeamMembers(org, teamSlug)` with pagination |
| `internal/api/enterprise_team.go` | `ListEnterpriseTeamMembers(enterprise, teamSlug)` with pagination |
| `internal/batch/runner.go` | `Run(ctx, client, enterprise, entries, concurrency, dryRun)` returns `BatchResult` |
| `internal/csv/parser.go` | `Parse(reader) -> []UserBudgetEntry` |
| `internal/output/formatter.go` | `PrintBudgets(writer, budgets, format)` |

---

## Verification

1. **Unit tests**: `go test ./... -race -cover` — target >80% coverage on `internal/` packages
2. **Manual smoke test**: `gh extension install .` locally, then:
   - `gh ulb set-universal -e test-enterprise -a 10.00` → verify budget created via `gh ulb list`
   - `gh ulb set-user -e test-enterprise -u octocat -a 50.00` → verify user override
   - `gh ulb set-team -e test-enterprise -o my-org -t my-team -a 25.00 --dry-run` → verify member list output
   - `gh ulb set-csv -e test-enterprise -f testdata/valid.csv --dry-run` → verify CSV parsing
   - `gh ulb delete -e test-enterprise -b BUDGET_ID --confirm` → verify deletion
3. **Lint**: `golangci-lint run` passes clean
4. **Build**: `go build -o gh-ulb` succeeds on macOS/Linux/Windows targets
5. **Release workflow**: Tag push triggers `gh-extension-precompile` and produces binaries

---

## Decisions

- **Go over Bash**: Better testing, cross-platform binaries, industry standard for `gh` extensions
- **Cobra CLI framework**: Matches `gh` itself, provides subcommands, help generation, flag parsing, shell completions
- **`go-gh` v2**: Handles auth (uses `gh` login token), host resolution, and REST client — no need to manage PATs manually in the extension
- **Concurrency model**: Bounded goroutine pool for batch operations (default 5) with exponential backoff + retry on 429/5xx responses (max 3 retries, jittered backoff starting at 1s). `--concurrency` flag lets users throttle for very large teams (1000+)
- **Idempotent batch updates**: Batch commands (set-team, set-enterprise-team, set-csv) attempt POST first; on 409 conflict (budget exists), automatically fall back to GET user budget → PATCH update. Each action is logged clearly: `[created]`, `[updated]`, or `[failed]` per user. This makes commands safe to re-run
- **Logging**: All mutating operations log the action taken per user to stdout (e.g. `✓ octocat: budget set to $50.00 [created]`, `✓ octocat: budget updated to $50.00 [updated]`, `✗ baduser: 404 not found [failed]`). Batch operations print a summary at the end (e.g. `3 created, 1 updated, 1 failed out of 5 users`)
- **API version**: `X-GitHub-Api-Version: 2022-11-28` as specified in the ULB Private Preview Guide (the billing budgets API uses this version)
- **Scope**: CRUD + batch. Alerting fields (`budget_alerting`) are stubbed since the feature isn't available yet (per FAQ). No UI — API-only during private preview
- **Excluded**: Budget reporting/dashboards, alerting configuration (not yet available), org-level budgets (enterprise-only feature)

---

## Agent Parallelization

The following agent assignments allow parallel implementation:

| Agent | Phase | Can Start | Work |
|---|---|---|---|
| **Agent A** | Phase 1 | Immediately | Scaffold, go.mod, root command, API client, budget CRUD, output formatter |
| **Agent F** | Phase 6 | Immediately | README, docs, examples, CI/CD workflows, LICENSE |
| **Agent B** | Phase 2 | After A | CRUD commands: set-universal, set-user, list, get, update, delete |
| **Agent C** | Phase 3 | After A | Team resolution (org + enterprise), batch runner |
| **Agent D** | Phase 4 | After B + C | Batch commands: set-team, set-enterprise-team, CSV parser, set-csv |
| **Agent E** | Phase 5 | After D | All unit tests across internal packages |

**Parallelism**: Agents A + F start in parallel. B + C start after A finishes. D starts after B + C finish. E starts after D. F runs fully in parallel with everything.

---

## Further Considerations

1. **CSV column flexibility**: Support `username`/`user`/`login` as column header variants via case-insensitive matching, to be forgiving with user-created CSVs.
