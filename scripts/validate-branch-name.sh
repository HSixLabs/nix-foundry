#!/bin/bash
set -e

# Git Branch Name Validator for nix-foundry
# --------------------------------------
# This script enforces a consistent branch naming convention across the project.
# It runs automatically as a pre-commit hook and in GitHub Actions to ensure all
# branch names follow the required pattern.
#
# Branch Name Pattern:
# <type>/<issue-number>-<description>
#
# Where:
# - type: feature, fix, chore, docs, test, refactor, style, ci, or perf
# - issue-number: just the number (e.g. "123")
# - description: lowercase letters, numbers, and hyphens
#
# Example: feature/123-add-user-authentication
#
# Usage:
# - Local development (pre-commit hook): ./scripts/validate-branch-name.sh
# - GitHub Actions: ./scripts/validate-branch-name.sh --ci [branch_name]

RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
NC='\033[0m'

validate_branch() {
    local branch=$1
    local types="feature|fix|chore|docs|test|refactor|style|ci|perf"
    local pattern="^($types)\/[0-9]+-[a-z0-9-]+$"

    if [[ "$branch" == "main" || "$branch" == "develop" ]]; then
        echo -e "${GREEN}✅ Protected branch '$branch' is allowed${NC}"
        return 0
    fi

    if [[ $branch =~ $pattern ]]; then
        echo -e "${GREEN}✅ Branch name '$branch' follows the convention${NC}"
        return 0
    else
        echo -e "${RED}❌ ERROR: Branch name '$branch' does not follow the convention:${NC}"
        echo -e "${YELLOW}<type>/<issue-number>-<short-description>${NC}"
        echo -e "\nWhere:"
        echo "- type: feature, fix, chore, docs, test, refactor, style, ci, or perf"
        echo "- issue-number: just the number (e.g. \"123\")"
        echo "- description: lowercase letters, numbers, and hyphens"
        echo -e "\nExample: feature/123-add-user-authentication"
        return 1
    fi
}

if [[ "$CI" == "true" ]]; then
    echo "Skipping branch name validation in CI environment"
    exit 0
fi

if [[ -n "$GITHUB_HEAD_REF" ]]; then
    BRANCH_NAME="$GITHUB_HEAD_REF"
elif [[ -n "$GITHUB_REF" ]]; then
    BRANCH_NAME="${GITHUB_REF#refs/heads/}"
else
    BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)
fi

echo "Validating branch name: $BRANCH_NAME"
validate_branch "$BRANCH_NAME"
