#!/bin/bash
set -e

cd "$(dirname "$0")"

RESOURCE_TYPE="checkmk_user"
RESOURCE_NAME="test"
RESOURCE_ID="lifecycle_test_user"

source ../run_lifecycle_test.sh
run_lifecycle_test "$RESOURCE_TYPE" "$RESOURCE_NAME" "$RESOURCE_ID"
