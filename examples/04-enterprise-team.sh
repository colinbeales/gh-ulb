#!/usr/bin/env bash
# Set budgets for an enterprise team
gh ulb set-enterprise-team \
  --enterprise my-enterprise \
  --team platform-engineers \
  --amount 30.00 \
  --dry-run

gh ulb set-enterprise-team \
  --enterprise my-enterprise \
  --team platform-engineers \
  --amount 30.00
