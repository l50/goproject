#!/bin/bash

# Pinned go-licenses module. v1 misreports stdlib packages when Go
# auto-switches to the toolchain in go.mod, so v2 is required.
GO_LICENSES="github.com/google/go-licenses/v2@v2.0.1"

# Run the pinned go-licenses rather than whatever is on PATH.
go_licenses() {
    go run "${GO_LICENSES}" "$@"
}

# Function to check if go mod vendor should run or not
run_vendor() {
    echo "Running go mod vendor..."
    go mod vendor
}

# Function to check licenses
check_licenses() {
    action=$1

    # Decide action based on input
    if [[ $action == "check_forbidden" ]]; then
        echo "Checking for forbidden licenses..."
        # go-licenses exits 1 both for disallowed licenses and for tool
        # failures, so tell them apart by its output.
        if output=$(go_licenses check ./... 2>&1); then
            echo "No forbidden licenses found."
        elif grep -qiE "not allowed license|did not find license for library" <<< "${output}"; then
            echo "${output}"
            echo "Forbidden licenses found. Please remove them."
            exit 1
        else
            echo "${output}" >&2
            echo "go-licenses failed to run; see the error above." >&2
            exit 1
        fi
    elif [[ $action == "output_csv" ]]; then
        echo "Outputting licenses to csv..."
        go_licenses csv ./...
    elif [[ $action == "vendor" ]]; then
        echo "Vendoring dependencies..."
        run_vendor
    fi
}

# Ensure input is provided
if [[ $# -lt 1 ]]; then
    echo "Incorrect number of arguments."
    echo "Usage: $0 <licenses action>"
    echo "Example: $0 check_forbidden"
    exit 1
fi

# Run checks
check_licenses "${1}"
