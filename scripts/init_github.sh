#!/usr/bin/env bash
set -euo pipefail
USER_NAME="${1:-your-github-username}"
REPO_NAME="${2:-quran-go}"
git init
git add .
git commit -m "chore: bootstrap quran-go monorepo"
git branch -M main
git remote add origin "git@github.com:${USER_NAME}/${REPO_NAME}.git"
git push -u origin main
