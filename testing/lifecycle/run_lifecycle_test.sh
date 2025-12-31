#!/bin/bash
# Generic lifecycle test runner
# Usage: source this file and call run_lifecycle_test

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_step() {
    echo -e "\n${YELLOW}==== $1 ====${NC}\n"
}

log_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

log_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Verify prerequisites
check_prerequisites() {
    if [ -z "$CHECKMK_URL" ]; then
        log_error "CHECKMK_URL not set. See testing/lifecycle/README.md for setup."
        exit 1
    fi

    if ! command -v terraform &> /dev/null; then
        log_error "terraform not found in PATH"
        exit 1
    fi

    if ! command -v curl &> /dev/null; then
        log_error "curl not found in PATH"
        exit 1
    fi
}

# Main lifecycle test function
# Args: $1 = resource_type (e.g., "checkmk_folder")
#       $2 = resource_name (e.g., "test")
#       $3 = resource_id for import (e.g., "my_folder")
run_lifecycle_test() {
    local resource_type="$1"
    local resource_name="$2"
    local resource_id="$3"
    local resource_address="${resource_type}.${resource_name}"

    check_prerequisites

    echo "========================================"
    echo "Lifecycle Test: $resource_type"
    echo "========================================"
    echo "Resource Address: $resource_address"
    echo "Import ID: $resource_id"
    echo "CheckMK URL: $CHECKMK_URL"
    echo "========================================"

    # Cleanup any previous state
    rm -rf .terraform terraform.tfstate* *.tfplan

    # Step 1: Init
    log_step "Step 1: Terraform Init"
    terraform init
    log_success "Init completed"

    # Step 2: Create
    log_step "Step 2: Create Resource"

    # Restore original main.tf if we have a backup
    if [ -f "main.tf.orig" ]; then
        cp main.tf.orig main.tf
    else
        cp main.tf main.tf.orig
    fi

    terraform plan -out=create.tfplan
    terraform apply create.tfplan
    log_success "Create completed"

    # Step 3: Refresh and verify no changes
    log_step "Step 3: Refresh and Verify State"
    terraform refresh

    PLAN_OUTPUT=$(terraform plan -detailed-exitcode 2>&1) || EXIT_CODE=$?
    if [ "${EXIT_CODE:-0}" -eq 0 ]; then
        log_success "No changes detected (state matches remote)"
    elif [ "${EXIT_CODE:-0}" -eq 2 ]; then
        log_error "Unexpected changes detected after create!"
        echo "$PLAN_OUTPUT"
        exit 1
    else
        log_error "Plan failed"
        echo "$PLAN_OUTPUT"
        exit 1
    fi

    # Step 4: Update
    log_step "Step 4: Update Resource"

    if [ -f "main_updated.tf" ]; then
        cp main_updated.tf main.tf
        terraform plan -out=update.tfplan
        terraform apply update.tfplan
        log_success "Update completed"

        # Verify no drift after update
        terraform refresh
        PLAN_OUTPUT=$(terraform plan -detailed-exitcode 2>&1) || EXIT_CODE=$?
        if [ "${EXIT_CODE:-0}" -eq 0 ]; then
            log_success "No changes detected after update"
        elif [ "${EXIT_CODE:-0}" -eq 2 ]; then
            log_error "Unexpected changes detected after update!"
            echo "$PLAN_OUTPUT"
            exit 1
        fi
    else
        echo "No main_updated.tf found, skipping update test"
    fi

    # Step 5: Import
    log_step "Step 5: Import Test"

    # Remove from state
    terraform state rm "$resource_address" || true

    # Import
    terraform import "$resource_address" "$resource_id"
    log_success "Import completed"

    # Verify import matches config
    PLAN_OUTPUT=$(terraform plan -detailed-exitcode 2>&1) || EXIT_CODE=$?
    if [ "${EXIT_CODE:-0}" -eq 0 ]; then
        log_success "Import state matches configuration"
    elif [ "${EXIT_CODE:-0}" -eq 2 ]; then
        log_error "State drift after import!"
        echo "$PLAN_OUTPUT"
        echo ""
        echo "This may indicate an issue with ImportState or Read."
        exit 1
    fi

    # Step 6: Destroy
    log_step "Step 6: Destroy Resource"
    terraform destroy -auto-approve
    log_success "Destroy completed"

    # Cleanup
    rm -f *.tfplan
    if [ -f "main.tf.orig" ]; then
        mv main.tf.orig main.tf
    fi

    echo ""
    echo "========================================"
    log_success "All lifecycle tests passed for $resource_type"
    echo "========================================"
}
