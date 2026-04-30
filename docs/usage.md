# Extended Usage Guide

## How budgets work

Copilot premium requests (e.g. GPT-4o, Claude Sonnet, o1) consume from a monthly budget that resets on the first day of each calendar month. Budgets are denominated in USD and track premium-only spend — standard Copilot requests included in your subscription are not counted.

## Hard vs soft budgets (`--prevent-overage`)

`gh-ulb` defaults to hard-stop behavior for budget commands:

- Hard budget (`--prevent-overage=true`, default): blocks further premium requests when the budget is exhausted.
- Soft budget (`--prevent-overage=false`): allows premium usage to continue after exhaustion while usage is tracked and reported.

This option is available on `set-universal`, `set-user`, `set-team`, `set-enterprise-team`, and `set-csv`.
For `update`, the existing behavior is kept unless `--prevent-overage` is explicitly provided.

When a budget is exhausted:

- If `--prevent-overage=true`, further premium requests are blocked for that user until the budget resets.
- If `--prevent-overage=false`, usage continues but is tracked and reported.

---

## Budget hierarchy

The API supports two types of budgets:

| Type | `target` field | Description |
|------|---------------|-------------|
| Universal | `multi_user_customer` | Applies to every Copilot user in the enterprise who does not have a personal override |
| Per-user | `user:<login>` | Overrides the universal budget for a specific user |

A per-user budget always wins over the universal budget. If a user has no personal budget, the universal budget applies (if one is set). If neither exists, there is no cap.

---

## When to use each command

| Command | Use case |
|---------|----------|
| `set-universal` | Apply a single spending cap to all users at once |
| `set-user` | Give a power user a higher (or lower) limit than the universal budget |
| `set-team` | Roll out budgets to everyone on an org-level GitHub Team |
| `set-enterprise-team` | Roll out budgets to everyone on an Enterprise Team |
| `set-csv` | Bulk-import budgets from a spreadsheet or HR data export |
| `list` | Audit all active budgets |
| `get` | Check what limit a specific user is operating under |
| `update` | Adjust an existing budget without deleting and recreating it |
| `delete` | Remove a budget (user reverts to universal budget or no limit) |
| `delete-universal` | Remove the universal enterprise-wide budget |
| `delete-team` | Remove budgets for all members of an org-level GitHub Team |
| `delete-enterprise-team` | Remove budgets for all members of an Enterprise Team |
| `delete-csv` | Remove budgets for all users listed in a CSV file |

---

## CSV import workflow

1. Export usernames from your IdP, HR system, or GitHub org membership list.
2. Create a CSV with a `username` (or `user`/`login`) column and an optional `amount` column.
3. Dry-run to validate:
   ```bash
   gh ulb set-csv --enterprise my-enterprise --file users.csv --dry-run
   ```
4. Review the preview output for any users that would fail (404 = not in enterprise).
5. Apply:
   ```bash
   gh ulb set-csv --enterprise my-enterprise --file users.csv
   ```

**CSV validation rules**

- The header row must contain a recognized username column (`username`, `user`, or `login` — case-insensitive).
- The `amount` column, if present, must be a valid decimal number (e.g. `25.00`).
- Rows with an empty username are skipped with a warning.
- If the `amount` column is absent, `--amount` is required on the command line.

---

## Enterprise team vs org team

| Dimension | Org team (`set-team`) | Enterprise team (`set-enterprise-team`) |
|-----------|----------------------|----------------------------------------|
| Scope | Single organization | Entire enterprise (cross-org) |
| API | `GET /orgs/{org}/teams/{team}/members` | `GET /enterprises/{enterprise}/teams/{team}/members` |
| Slug source | Team slug in the org | Enterprise team slug |

Use `set-enterprise-team` when your team spans multiple organizations within the enterprise.

---

## Re-running batch commands safely

All batch commands (`set-team`, `set-enterprise-team`, `set-csv`) are **idempotent**:

- If a budget already exists for a user (`409 Conflict`), the command automatically updates the existing budget instead of failing.
- If a user is not found in the enterprise (`404 Not Found`), the error is logged per-user and the command continues with remaining users.
- Exit code is non-zero only if _all_ users fail; partial success returns exit code 0 with per-user warnings printed to stderr.

This means you can safely re-run batch commands as part of a scheduled job or GitOps workflow.

---

## Common patterns

### Rolling out budgets to a team

```bash
# 1. Preview
gh ulb set-team -e my-enterprise -o my-org -t engineering --amount 25.00 --dry-run

# 2. Apply
gh ulb set-team -e my-enterprise -o my-org -t engineering --amount 25.00
```

### Updating limits mid-month

Use `list` to find the budget IDs, then `update`:

```bash
gh ulb list -e my-enterprise --json | jq '.[] | select(.target == "user:octocat") | .id'
gh ulb update -e my-enterprise --budget-id <id> --amount 100.00
```

### Removing a budget

```bash
# Interactive
gh ulb delete -e my-enterprise --budget-id <id>

# Scripted
gh ulb delete -e my-enterprise --budget-id <id> --confirm
```

After deletion the user reverts to the universal budget (if set) or has no cap.

### Auditing spend

```bash
gh ulb list -e my-enterprise --json | jq 'sort_by(.amount) | reverse'
```
