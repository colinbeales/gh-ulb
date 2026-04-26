#!/usr/bin/env bash
# Delete the enterprise-wide universal budget (multi_user_customer scope)

# Interactive confirmation
gh ulb delete-universal-budget \
  --enterprise my-enterprise

# Non-interactive / scripting
gh ulb delete-universal-budget \
  --enterprise my-enterprise \
  --confirm
