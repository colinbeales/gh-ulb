# gh-ulb — GitHub CLI Extension for Copilot User-Level Budgets

![Go Version](https://img.shields.io/badge/go-1.22-blue)
![License](https://img.shields.io/badge/license-MIT-green)

`gh-ulb` is a GitHub CLI extension for managing Copilot **User-Level Budgets (ULB)** — enterprise-level controls over Copilot spend. With ULB you can set monthly spending caps per user, override them for specific individuals, and manage budgets in bulk via GitHub Teams or CSV import.

---

## Prerequisites

- [`gh` CLI](https://cli.github.com/) installed and authenticated (`gh auth login`)
- Enterprise admin permissions
- Copilot Business or Enterprise subscription with premium requests enabled

### Authentication Scopes

This extension requires enterprise billing and team-read scopes.

For most users, these classic PAT scopes are sufficient:

- `manage_billing:enterprise`, `read:org`, `read:enterprise`
- or `admin:enterprise` + `read:org`

For full command-by-command requirements, scope verification, and refresh commands, see [docs/permissions.md](docs/permissions.md).

---

## Installation

```bash
gh extension install colinbeales/gh-ulb
```

## Working with GitHub Enterprise Cloud with Data Residency

For GitHub Enterprise Cloud with Data Residency enterprises hosted on GHE.com, API calls must target your enterprise host (for example, `api.<subdomain>.ghe.com`) instead of `api.github.com`.

`gh-ulb` supports this via either:

- `--hostname <subdomain>.ghe.com` on any command
- `GH_HOST=<subdomain>.ghe.com` environment variable

Examples:

```bash
gh ulb list --enterprise my-enterprise --hostname my-enterprise.ghe.com

GH_HOST=my-enterprise.ghe.com
gh ulb set-enterprise-team \
  --enterprise my-enterprise \
  --team foo \
  --amount 15.00
```

Important: authenticate `gh` against the same host before running commands:

```bash
gh auth login --hostname my-enterprise.ghe.com
```

If the auth host and command host differ, requests may fail with 401/403/404 even when flags and scopes look correct.

---

## Commands

All commands accept a `--enterprise` / `-e` flag (required) to specify the GitHub Enterprise slug.

### Hard vs Soft Budgets

`gh-ulb` is opinionated toward hard budgets by default:

- `--prevent-overage=true` (default): hard-stop. Requests using AI Credits are blocked after the budget is reached.
- `--prevent-overage=false`: soft budget. Usage can continue past budget alert.

You can override this on `set-universal`, `set-user`, `set-team`, `set-enterprise-team`, `set-csv`, and `update`.

### `set-universal`

Create or update the universal budget that applies to all Copilot users in the enterprise.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--amount` | `-a` | Monthly budget in USD (required) |
| `--prevent-overage` | | Block usage once budget is reached (default: `true`; set `false` for soft budget) |
| `--dry-run` | | Preview without making changes |
| `--json` | | Output result as JSON |

**Example**

```bash
gh ulb set-universal \
  --enterprise my-enterprise \
  --amount 10.00 \
  --prevent-overage
```

---

### `set-user`

Create a per-user budget override for a specific enterprise member.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--user` | `-u` | GitHub username (required) |
| `--amount` | `-a` | Monthly budget in USD (required) |
| `--prevent-overage` | | Block usage once budget is reached (default: `true`; set `false` for soft budget) |
| `--dry-run` | | Preview without making changes |
| `--json` | | Output result as JSON |

**Example**

```bash
gh ulb set-user \
  --enterprise my-enterprise \
  --user octocat \
  --amount 50.00
```

---

### `set-team`

Set budgets for all members of an organization-level GitHub Team.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--org` | `-o` | Organization slug (required) |
| `--team` | `-t` | Team slug (required) |
| `--amount` | `-a` | Monthly budget in USD (required) |
| `--prevent-overage` | | Block usage once budget is reached (default: `true`; set `false` for soft budget) |
| `--concurrency` | `-c` | Number of concurrent API calls (default: 5) |
| `--dry-run` | | Preview without making changes |
| `--json` | | Output result as JSON |

**Example**

```bash
gh ulb set-team \
  --enterprise my-enterprise \
  --org my-org \
  --team engineering \
  --amount 25.00 \
  --concurrency 10
```

---

### `set-enterprise-team`

Set budgets for all members of an Enterprise Team (not org-scoped).

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--team` | `-t` | Enterprise team slug (required) |
| `--amount` | `-a` | Monthly budget in USD (required) |
| `--prevent-overage` | | Block usage once budget is reached (default: `true`; set `false` for soft budget) |
| `--concurrency` | `-c` | Number of concurrent API calls (default: 5) |
| `--dry-run` | | Preview without making changes |
| `--json` | | Output result as JSON |

**Example**

```bash
gh ulb set-enterprise-team \
  --enterprise my-enterprise \
  --team platform-engineers \
  --amount 30.00
```

---

### `set-csv`

Set budgets for multiple users from a CSV file.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--file` | `-f` | Path to CSV file (required) |
| `--amount` | `-a` | Default monthly budget if CSV has no `amount` column |
| `--prevent-overage` | | Block usage once budget is reached (default: `true`; set `false` for soft budget) |
| `--concurrency` | `-c` | Number of concurrent API calls (default: 5) |
| `--dry-run` | | Preview without making changes |
| `--json` | | Output result as JSON |

The [CSV file format](#csv-format) must have a username (or user/login) column. An amount column is not required. If not present `--amount` flag is required and will be used for all users.

**Example**

```bash
gh ulb set-csv \
  --enterprise my-enterprise \
  --file users.csv \
  --dry-run

gh ulb set-csv \
  --enterprise my-enterprise \
  --file users.csv \
  --prevent-overage=false
```

---

### `list`

List all budgets configured for the enterprise.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--json` | | Output as JSON |

**Example**

```bash
gh ulb list --enterprise my-enterprise --json
```

---

### `get`

Get the effective budget for a specific user.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--user` | `-u` | GitHub username (required) |
| `--json` | | Output as JSON |

**Example**

```bash
gh ulb get --enterprise my-enterprise --user octocat
```

---

### `update`

Update an existing budget by its ID.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--budget-id` | `-b` | Budget ID to update (required) |
| `--amount` | `-a` | New monthly budget in USD (required) |
| `--prevent-overage` | | Block usage once budget is reached (default: keep existing value unless explicitly set) |
| `--json` | | Output as JSON |

**Example**

```bash
gh ulb update \
  --enterprise my-enterprise \
  --budget-id BUDGET_ID \
  --amount 75.00
```

---

### `delete`

Delete a budget by its ID.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--budget-id` | `-b` | Budget ID to delete (required) |
| `--confirm` | | Skip confirmation prompt |

**Example**

```bash
# Interactive confirmation
gh ulb delete --enterprise my-enterprise --budget-id BUDGET_ID

# Non-interactive / scripting
gh ulb delete --enterprise my-enterprise --budget-id BUDGET_ID --confirm
```

---

### `delete-universal`

Delete the universal budget (`multi_user_customer` scope).

This command lists enterprise budgets, finds the `multi_user_customer` entry, and deletes it.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--confirm` | | Skip confirmation prompt |

**Example**

```bash
# Interactive confirmation
gh ulb delete-universal --enterprise my-enterprise

# Non-interactive / scripting
gh ulb delete-universal --enterprise my-enterprise --confirm
```

---

### `delete-team`

Delete per-user budgets for all members of an organization-level GitHub Team.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--org` | `-o` | Organization slug (required) |
| `--team` | `-t` | Team slug (required) |
| `--concurrency` | | Number of concurrent API calls (default: 5) |
| `--dry-run` | | Preview without making changes |
| `--confirm` | | Skip confirmation prompt |

**Example**

```bash
# Preview
gh ulb delete-team \
  --enterprise my-enterprise \
  --org my-org \
  --team engineering \
  --dry-run

# Apply
gh ulb delete-team \
  --enterprise my-enterprise \
  --org my-org \
  --team engineering \
  --confirm
```

---

### `delete-enterprise-team`

Delete per-user budgets for all members of an Enterprise Team (not org-scoped).

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--team` | `-t` | Enterprise team slug (required) |
| `--concurrency` | | Number of concurrent API calls (default: 5) |
| `--dry-run` | | Preview without making changes |
| `--confirm` | | Skip confirmation prompt |

**Example**

```bash
# Preview
gh ulb delete-enterprise-team \
  --enterprise my-enterprise \
  --team platform-engineers \
  --dry-run

# Apply
gh ulb delete-enterprise-team \
  --enterprise my-enterprise \
  --team platform-engineers \
  --confirm
```

---

### `delete-csv`

Delete per-user budgets for all users listed in a CSV file.

**Flags**

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | Enterprise slug (required) |
| `--file` | `-f` | Path to CSV file (required) |
| `--concurrency` | | Number of concurrent API calls (default: 5) |
| `--dry-run` | | Preview without making changes |
| `--confirm` | | Skip confirmation prompt |

The [CSV file format](#csv-format) must have a `username` (or `user`/`login`) column. An `amount` column is not required and will be ignored.

**Example**

```bash
# Preview
gh ulb delete-csv \
  --enterprise my-enterprise \
  --file offboarded-users.csv \
  --dry-run

# Apply
gh ulb delete-csv \
  --enterprise my-enterprise \
  --file offboarded-users.csv \
  --confirm
```

---

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--enterprise` | `-e` | GitHub Enterprise slug |
| `--json` | | Output as JSON instead of human-readable table |

---

## CSV Format

The CSV file must have a header row. Supported username column names (case-insensitive): `username`, `user`, or `login`.

The `amount` column is optional. If omitted, the `--amount` flag value is used as the default for every row.

```csv
username,amount
octocat,50.00
monalisa,25.00
hubot,10.00
```

Username-only example (requires `--amount` flag):

```csv
username
octocat
monalisa
hubot
```

---

## Budget Hierarchy

| Type | API target | Scope |
|------|-----------|-------|
| Universal (`multi_user_customer`) | All users in enterprise | Lowest priority |
| Per-user | Specific login | Overrides universal |

A per-user budget always takes precedence over the universal budget for that user.

---

## Troubleshooting

| Symptom | Cause | Resolution |
|---------|-------|-----------|
| `404 Not Found` | User is not a member of the enterprise | Verify the username and enterprise membership |
| `409 Conflict` | A budget for this user already exists | Batch commands (`set-team`, `set-csv`) handle 409s automatically by updating the existing budget |
| Rate limiting (`429`) | Too many concurrent API requests | Reduce `--concurrency`; batch commands retry automatically with exponential backoff |
| Unexpected changes | Want to preview before applying | Use `--dry-run` on any batch command to see what would happen without making API calls |

---

## Development & Releases

For maintainer release procedures, see [docs/releasing.md](docs/releasing.md).

---

## License

MIT — see [LICENSE](LICENSE).
