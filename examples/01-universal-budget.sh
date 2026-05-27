#!/usr/bin/env bash
# Set a $10/month universal budget for all Copilot users
gh ulb set-universal \
  --enterprise my-enterprise \
  --amount 10.00

# List all budgets to confirm
gh ulb list --enterprise my-enterprise

# Remove the universal budget again (interactive)
gh ulb delete-universal --enterprise my-enterprise

# Remove without prompt (for scripting)
gh ulb delete-universal --enterprise my-enterprise --confirm
