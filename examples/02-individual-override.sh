#!/usr/bin/env bash
# Give a specific power user a higher budget
gh ulb set-user \
  --enterprise my-enterprise \
  --user octocat \
  --amount 50.00

# Verify the override
gh ulb get --enterprise my-enterprise --user octocat
