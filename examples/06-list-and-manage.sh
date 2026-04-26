#!/usr/bin/env bash
# List all budgets
gh ulb list --enterprise my-enterprise

# List as JSON for scripting
gh ulb list --enterprise my-enterprise --json

# Get a specific user's effective budget
gh ulb get --enterprise my-enterprise --user octocat

# Update a budget by ID
gh ulb update --enterprise my-enterprise --budget-id BUDGET_ID --amount 75.00

# Delete a budget (with confirmation prompt)
gh ulb delete --enterprise my-enterprise --budget-id BUDGET_ID

# Delete without prompt (for scripting)
gh ulb delete --enterprise my-enterprise --budget-id BUDGET_ID --confirm
