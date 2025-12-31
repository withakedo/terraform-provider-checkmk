#!/bin/bash
set -e

# Test Rule API - uses environment variables for credentials
#
# Required environment variables:
#   CHECKMK_URL      - e.g., http://localhost:5030/test
#   CHECKMK_USERNAME - e.g., automation
#   CHECKMK_PASSWORD - retrieved from container
#
# Usage:
#   ./scripts/run_tests.sh  # Sets up credentials automatically
#   # Or manually (get password from container env):
#   export CHECKMK_URL=http://localhost:5030/test
#   export CHECKMK_USERNAME=automation
#   export CHECKMK_PASSWORD=$(docker inspect --format '{{range .Config.Env}}{{println .}}{{end}}' checkmk-2.3.0p41 | grep CMK_AUTOMATION_SECRET | cut -d= -f2-)
#   ./testing/test_rule_api.sh

if [ -z "$CHECKMK_URL" ] || [ -z "$CHECKMK_USERNAME" ] || [ -z "$CHECKMK_PASSWORD" ]; then
    echo "ERROR: Required environment variables not set"
    echo "  CHECKMK_URL, CHECKMK_USERNAME, CHECKMK_PASSWORD"
    echo ""
    echo "Use ./scripts/run_tests.sh to set credentials automatically"
    exit 1
fi

API_BASE="${CHECKMK_URL}/check_mk/api/1.0"
AUTH="${CHECKMK_USERNAME}:${CHECKMK_PASSWORD}"

echo "=== Testing Rule API ==="
echo "API: $API_BASE"
echo

# Test 1: Create a rule
echo "1. Creating a rule..."
CREATE_RESPONSE=$(curl -s -X POST -u "$AUTH" \
  -H "Content-Type: application/json" \
  "$API_BASE/domain-types/rule/collections/all" \
  -d '{
    "ruleset": "host_label_rules",
    "folder": "/",
    "properties": {
      "description": "Test rule from API test",
      "comment": "Testing the rule API",
      "disabled": false
    },
    "value_raw": "{\"test_label\": \"test_value\"}",
    "conditions": {}
  }')

RULE_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')
echo "Created rule with ID: $RULE_ID"
echo

# Test 2: Get the rule
echo "2. Getting the rule..."
GET_RESPONSE=$(curl -s -u "$AUTH" "$API_BASE/objects/rule/$RULE_ID")
DESCRIPTION=$(echo "$GET_RESPONSE" | jq -r '.extensions.properties.description')
echo "Description: $DESCRIPTION"
echo

# Test 3: Get ETag
echo "3. Getting ETag..."
ETAG=$(curl -s -i -u "$AUTH" "$API_BASE/objects/rule/$RULE_ID" | grep -i '^etag:' | awk '{print $2}' | tr -d '\r')
echo "ETag: $ETAG"
echo

# Test 4: List rules in ruleset
echo "4. Listing rules in ruleset..."
LIST_RESPONSE=$(curl -s -u "$AUTH" "$API_BASE/domain-types/rule/collections/all?ruleset_name=host_label_rules")
COUNT=$(echo "$LIST_RESPONSE" | jq '.value | length')
echo "Found $COUNT rules in host_label_rules"
echo

# Test 5: Update the rule
echo "5. Updating the rule..."
UPDATE_RESPONSE=$(curl -s -X PUT -u "$AUTH" \
  -H "Content-Type: application/json" \
  -H "If-Match: $ETAG" \
  "$API_BASE/objects/rule/$RULE_ID" \
  -d '{
    "properties": {
      "description": "Updated test rule",
      "comment": "This rule was updated",
      "disabled": false
    },
    "value_raw": "{\"test_label\": \"updated_value\"}",
    "conditions": {}
  }')

NEW_DESCRIPTION=$(echo "$UPDATE_RESPONSE" | jq -r '.extensions.properties.description')
echo "New description: $NEW_DESCRIPTION"
echo

# Test 6: Get new ETag
echo "6. Getting new ETag..."
NEW_ETAG=$(curl -s -i -u "$AUTH" "$API_BASE/objects/rule/$RULE_ID" | grep -i '^etag:' | awk '{print $2}' | tr -d '\r')
echo "New ETag: $NEW_ETAG"
echo

# Test 7: Delete the rule
echo "7. Deleting the rule..."
DELETE_RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE -u "$AUTH" \
  -H "If-Match: $NEW_ETAG" \
  "$API_BASE/objects/rule/$RULE_ID")
DELETE_STATUS=$(echo "$DELETE_RESPONSE" | tail -n 1)
echo "Delete status: $DELETE_STATUS"
echo

# Test 8: Verify deletion
echo "8. Verifying deletion..."
VERIFY_RESPONSE=$(curl -s -w "\n%{http_code}" -u "$AUTH" "$API_BASE/objects/rule/$RULE_ID")
VERIFY_STATUS=$(echo "$VERIFY_RESPONSE" | tail -n 1)
if [ "$VERIFY_STATUS" = "404" ]; then
  echo "Rule successfully deleted (404 Not Found)"
else
  echo "ERROR: Expected 404, got $VERIFY_STATUS"
  exit 1
fi
echo

echo "=== All API Tests Passed! ==="
