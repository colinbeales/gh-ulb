#!/usr/bin/env bash
# Delete budgets for users listed in a CSV file
#
# The CSV only needs a username column — amount is not required:
#   username
#   octocat
#   monalisa
#   hubot

# Preview which budgets would be removed (no changes made)
gh ulb delete-csv \
  --enterprise my-enterprise \
  --file offboarded-users.csv \
  --dry-run

# Apply — prompts for confirmation
gh ulb delete-csv \
  --enterprise my-enterprise \
  --file offboarded-users.csv

# Apply without prompt (for scripting / CI)
gh ulb delete-csv \
  --enterprise my-enterprise \
  --file offboarded-users.csv \
  --confirm
