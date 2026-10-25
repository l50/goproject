#!/bin/bash

REPO_URL=$(git config --get remote.origin.url)

if [[ "${REPO_URL}" == *https* ]]; then
    REPO=$(echo "${REPO_URL}" | awk -F '://' '{print $2}' | sed 's/\.git$//')
else
    REPO=$(echo "${REPO_URL}" | awk -F '@' '{print $2}' | tr : / | sed 's/\.git$//')
fi

if grep -F "replace ${REPO}" "$@" 2>&1; then
    echo "ERROR: Don't commit a replacement in go.mod!"
    exit 1
fi
