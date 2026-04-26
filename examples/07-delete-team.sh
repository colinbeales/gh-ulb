#!/usr/bin/env bash
# Delete budgets for all members of an org team

# Preview which budgets would be removed (no changes made)
gh ulb delete-team \
  --enterprise my-enterprise \
  --org my-org \
  --team engineering \
  --dry-run

# Apply — prompts for confirmation
gh ulb delete-team \
  --enterprise my-enterprise \
  --org my-org \
  --team engineering

# Apply without prompt (for scripting / CI)
gh ulb delete-team \
  --enterprise my-enterprise \
  --org my-org \
  --team engineering \
  --confirm
