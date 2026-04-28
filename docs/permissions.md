# Permissions and Authentication Scopes

`gh-ulb` calls enterprise billing and team membership APIs. Your `gh` authentication token must include the right scopes for the commands you run.

## Recommended Scope Sets (Classic PAT)

Use one of these approaches:

- Least privilege for all `gh ulb` commands: `manage_billing:enterprise`, `read:org`, `read:enterprise`
- Single broad scope alternative: `admin:enterprise` plus `read:org`

## Scope Requirements by Command

| Command | APIs used | Required classic PAT scopes |
|------|-------|-------------|
| `set-universal` | `POST /enterprises/{enterprise}/settings/billing/budgets` | `manage_billing:enterprise` (or `admin:enterprise`) |
| `set-user` | `POST /enterprises/{enterprise}/settings/billing/budgets` | `manage_billing:enterprise` (or `admin:enterprise`) |
| `set-csv` | `POST/GET/PATCH /enterprises/{enterprise}/settings/billing/budgets...` | `manage_billing:enterprise` (or `admin:enterprise`) |
| `list` | `GET /enterprises/{enterprise}/settings/billing/budgets` | `manage_billing:enterprise` (or `admin:enterprise`) |
| `get` | `GET /enterprises/{enterprise}/settings/billing/budgets?user=...` | `manage_billing:enterprise` (or `admin:enterprise`) |
| `update` | `PATCH /enterprises/{enterprise}/settings/billing/budgets/{id}` | `manage_billing:enterprise` (or `admin:enterprise`) |
| `delete` | `DELETE /enterprises/{enterprise}/settings/billing/budgets/{id}` | `manage_billing:enterprise` (or `admin:enterprise`) |
| `delete-universal-budget` | `GET /enterprises/{enterprise}/settings/billing/budgets` + `DELETE /enterprises/{enterprise}/settings/billing/budgets/{id}` | `manage_billing:enterprise` (or `admin:enterprise`) |
| `set-team` | `GET /orgs/{org}/teams/{team}/members` + budget APIs | `read:org` + `manage_billing:enterprise` (or `admin:enterprise`) |
| `set-enterprise-team` | `GET /enterprises/{enterprise}/teams/{team}/memberships` + budget APIs | `read:enterprise` + `manage_billing:enterprise` (or `admin:enterprise`) |
| `delete-team` | `GET /orgs/{org}/teams/{team}/members` + `DELETE` budget APIs | `read:org` + `manage_billing:enterprise` (or `admin:enterprise`) |
| `delete-enterprise-team` | `GET /enterprises/{enterprise}/teams/{team}/memberships` + `DELETE` budget APIs | `read:enterprise` + `manage_billing:enterprise` (or `admin:enterprise`) |
| `delete-csv` | `GET` + `DELETE` budget APIs | `manage_billing:enterprise` (or `admin:enterprise`) |

If you also run raw enterprise-team discovery calls like `GET /enterprises/{enterprise}/teams`, include `read:enterprise` (or use `admin:enterprise`).

## Verify Current Scopes

```bash
gh auth status -t
gh api -i /user | grep -iE '^(x-oauth-scopes|x-accepted-oauth-scopes):'
```

## Request Additional Scopes

```bash
gh auth refresh -h github.com -s manage_billing:enterprise -s read:enterprise -s read:org
```
