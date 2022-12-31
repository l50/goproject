#!/bin/bash
set -ex

pkg=$(go list ./...)
for dir in */; do
    if [[ "${dir}" != ".hooks" ]] \
                              && [[ "${dir}" != "bin/" ]] \
                              && [[ "${dir}" != "config/" ]] \
                              && [[ "${dir}" != "cmd/" ]] \
                              && [[ "${dir}" != "docs/" ]] \
                              && [[ "${dir}" != "dist/" ]] \
                              && [[ "${dir}" != "images/" ]] \
                              && [[ "${dir}" != "files/" ]] \
                              && [[ "${dir}" != "logs/" ]] \
                              && [[ "${dir}" != "magefiles/" ]] \
                              && [[ "${dir}" != "templates/" ]] \
                              && [[ "${dir}" != "resources/" ]]; then
        go vet "${pkg}/${dir}"
    fi
done
