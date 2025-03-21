#!/bin/bash

# Exit on error and pipe failure
set -e -o pipefail

BLUE="\033[0;34m"
YELLOW="\033[0;33m"
GREEN="\033[0;32m"
RED="\033[0;31m"
RESET="\033[0m"

# Install dependencies if missing
install_dependencies() {
    # Check for golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        echo -e "🔍 golangci-lint could not be found, installing..." ; echo
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi

    # Check for gosec
    if ! command -v gosec &> /dev/null; then
        echo -e "🔍 gosec could not be found, installing..." ; echo
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    fi
}

# Run linters with error handling
run_linters() {
    local lint_status=0
    local security_status=0

    echo -e "${BLUE}🔎 Running golangci-lint...${RESET}" ; echo
    golangci-lint -v run ./... || lint_status=$?
    if [ $lint_status -ne 0 ]; then
        echo -e "${RED}❌ Linting issues found:${RESET}" ; echo
        [ $lint_status -ne 0 ] && echo "- Go static analysis failed (golangci-lint)" ; echo
        exit 1
    fi
    echo -e "${GREEN}✅ Linting passed${RESET}" ; echo

    echo -e "${BLUE}🛡️  Running gosec (security analysis)...${RESET}" ; echo
    gosec ./... || security_status=$?

    if [ $security_status -ne 0 ]; then
        echo -e "${RED}❌ Security issues found:${RESET}" ; echo
        [ $security_status -ne 0 ] && echo "- Security analysis failed (gosec)" ; echo
        exit 1
    fi
    echo -e "${GREEN}✅ Security analysis passed${RESET}" ; echo
    return 0
}

function main() {
    install_dependencies
    # check_for_updates
    run_linters
    echo -e "${GREEN}🎉 All checks passed successfully${RESET}" ; echo
}

# Execute main function
main

