#!/bin/bash

# Function to check if go mod vendor should run or not
run_vendor() {
    echo "Running go mod vendor..."
    go mod vendor
}

# Function to check licenses
check_licenses() {
    action=$1

    # Only install if not available
    if ! command -v go-licenses &> /dev/null; then
        go install github.com/google/go-licenses@latest
    fi

    # Decide action based on input
    if [[ $action == "check_forbidden" ]]; then
        echo "Checking for forbidden licenses..."
        if ! go-licenses check ./...; then
            echo "Forbidden licenses found. Please remove them."
            exit 1
        else
            echo "No forbidden licenses found."
        fi
    elif [[ $action == "output_csv" ]]; then
        echo "Outputting licenses to csv..."
        go-licenses csv ./...
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
