#!/usr/bin/env bash
# Delete budgets for all members of an enterprise team

# Preview which budgets would be removed (no changes made)
gh ulb delete-enterprise-team \
  --enterprise my-enterprise \
  --team platform-engineers \
  --dry-run

# Apply — prompts for confirmation
gh ulb delete-enterprise-team \
  --enterprise my-enterprise \
  --team platform-engineers

# Apply without prompt (for scripting / CI)
gh ulb delete-enterprise-team \
  --enterprise my-enterprise \
  --team platform-engineers \
  --confirm
