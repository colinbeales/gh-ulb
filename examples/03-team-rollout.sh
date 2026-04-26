#!/usr/bin/env bash
# Preview which users would be affected (dry run)
gh ulb set-team \
  --enterprise my-enterprise \
  --org my-org \
  --team engineering \
  --amount 25.00 \
  --dry-run

# Apply the budget to all team members
gh ulb set-team \
  --enterprise my-enterprise \
  --org my-org \
  --team engineering \
  --amount 25.00 \
  --concurrency 10
