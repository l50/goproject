#!/bin/bash
set -ex

pkgs=$(go list ./...)

for pkg in $pkgs; do
    dir="$(basename "$pkg")/"
    if [[ "${dir}" != ".hooks/" ]] \
                              && [[ "${dir}" != ".github/" ]] \
                              && [[ "${dir}" != "bin/" ]] \
                              && [[ "${dir}" != "config/" ]] \
                              && [[ "${dir}" != "deployments/" ]] \
                              && [[ "${dir}" != "docs/" ]] \
                              && [[ "${dir}" != "files/" ]] \
                              && [[ "${dir}" != "images/" ]] \
                              && [[ "${dir}" != "logs/" ]] \
                              && [[ "${dir}" != "magefiles/" ]] \
                              && [[ "${dir}" != "modules/" ]] \
                              && [[ "${dir}" != "resources/" ]] \
                              && [[ "${dir}" != "templates/" ]]; then
        go vet "${pkg}/${dir}"
    fi
done
