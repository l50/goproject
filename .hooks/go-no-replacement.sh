#!/bin/bash

REPO_URL=$(git config --get remote.origin.url)

if [[ "${REPO_URL}" == *https* ]]; then
    REPO=$(echo "${REPO_URL}" | awk -F '://' '{print $2}' | sed 's/\.git$//')
else
    REPO=$(echo "${REPO_URL}" | awk -F '@' '{print $2}' | tr : / | sed 's/\.git$//')
fi

# Match the repo path as the source of a replace directive, in both the
# single-line form and inside a replace ( ... ) block.
REPO_RE=$(printf '%s' "${REPO}" | sed 's/[][\.*^$()+?{}|]/\\&/g')
if grep -E "^[[:space:]]*(replace[[:space:]]+)?${REPO_RE}([[:space:]]+v[^[:space:]]+)?[[:space:]]+=>" "$@" 2>&1; then
    echo "ERROR: Don't commit a replacement in go.mod!"
    exit 1
fi
