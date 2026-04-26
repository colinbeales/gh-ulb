#!/usr/bin/env bash
# CSV format: username,amount
# Create a file users.csv:
#   username,amount
#   octocat,50.00
#   monalisa,25.00
#   hubot,10.00

# Dry run to validate
gh ulb set-csv \
  --enterprise my-enterprise \
  --file users.csv \
  --dry-run

# Apply all budgets
gh ulb set-csv \
  --enterprise my-enterprise \
  --file users.csv

# With a default amount (for rows without an amount column)
gh ulb set-csv \
  --enterprise my-enterprise \
  --file users-no-amount.csv \
  --amount 20.00
