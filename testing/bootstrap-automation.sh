#!/bin/bash
# Bootstrap automation user for testing
# This script runs after CheckMK is fully initialized (post-start hook)
# It creates the automation user and sets up credentials for API access

set -e

SITE_ID="${CMK_SITE_ID:-test}"
AUTOMATION_SECRET="${CMK_AUTOMATION_SECRET:-testSecret123}"
ADMIN_PASSWORD="${CMK_PASSWORD:-test123}"

echo "[BOOTSTRAP] Configuring automation user for site: $SITE_ID"

# Wait for site to be fully ready
echo "[BOOTSTRAP] Waiting for site to be ready..."
MAX_WAIT=60
WAITED=0
while ! omd status "$SITE_ID" >/dev/null 2>&1; do
    if [ $WAITED -ge $MAX_WAIT ]; then
        echo "[BOOTSTRAP] ERROR: Site did not become ready within ${MAX_WAIT}s"
        exit 1
    fi
    sleep 1
    WAITED=$((WAITED + 1))
done
echo "[BOOTSTRAP] Site is ready"

# Ensure automation user exists in users.mk
USERS_MK="/omd/sites/$SITE_ID/etc/check_mk/multisite.d/wato/users.mk"
if ! grep -q "'automation'" "$USERS_MK" 2>/dev/null; then
    echo "[BOOTSTRAP] Adding automation user to users.mk"
    cat >> "$USERS_MK" << 'EOF'

# Automation user for API access (added by bootstrap script)
multisite_users["automation"] = {
    "alias": "Automation User",
    "roles": ["admin"],
    "locked": False,
    "connector": "htpasswd",
}
EOF
    chown "$SITE_ID:$SITE_ID" "$USERS_MK"
fi

# Ensure automation secret directory and file exist
AUTOMATION_DIR="/omd/sites/$SITE_ID/var/check_mk/web/automation"
echo "[BOOTSTRAP] Creating automation secret directory: $AUTOMATION_DIR"
mkdir -p "$AUTOMATION_DIR"
echo "$AUTOMATION_SECRET" > "$AUTOMATION_DIR/automation.secret"
chown -R "$SITE_ID:$SITE_ID" "$AUTOMATION_DIR"
chmod 640 "$AUTOMATION_DIR/automation.secret"

# Add automation user to htpasswd (required for Basic auth)
# Use -B for bcrypt which CheckMK requires
echo "[BOOTSTRAP] Adding automation user to htpasswd"
htpasswd -Bb "/omd/sites/$SITE_ID/etc/htpasswd" automation "$AUTOMATION_SECRET"
chown "$SITE_ID:$SITE_ID" "/omd/sites/$SITE_ID/etc/htpasswd"

# Update cmkadmin password for web UI access
echo "[BOOTSTRAP] Setting cmkadmin password"
htpasswd -Bb "/omd/sites/$SITE_ID/etc/htpasswd" cmkadmin "$ADMIN_PASSWORD"
chown "$SITE_ID:$SITE_ID" "/omd/sites/$SITE_ID/etc/htpasswd"

echo "[BOOTSTRAP] Automation user configured successfully"
echo "[BOOTSTRAP] - API user: automation"
echo "[BOOTSTRAP] - Web UI user: cmkadmin"
echo "[BOOTSTRAP] Credentials available via: docker inspect --format '{{range .Config.Env}}{{println .}}{{end}}' <container>"
