# Testing Custom Attributes

This directory contains files for bootstrapping custom attributes into CheckMK test containers.

## Files

- `custom_attrs.mk` - Custom attribute definitions for testing
- `bootstrap-custom-attrs.sh` - Script to copy custom attributes into container

## Why This is Needed

Custom attributes in CheckMK:
- Can only be created via CLI (not through REST API)
- Are not listed in the OpenAPI specification
- Must exist in CheckMK BEFORE the Terraform provider can use them
- Will cause API errors if referenced but not configured

## Usage with Docker Containers

### Option 1: Docker Compose

```yaml
version: '3.8'

services:
  checkmk:
    image: checkmk/check-mk-enterprise:2.3.0p41
    volumes:
      # Mount custom attributes file
      - ./testing/custom_attrs.mk:/wato/custom_attrs.mk:ro
      # Mount bootstrap script
      - ./testing/bootstrap-custom-attrs.sh:/docker-entrypoint.d/pre-start/custom_attrs.sh:ro
    environment:
      - CMK_SITE_ID=cmk
      - CMK_PASSWORD=your-password-here
```

### Option 2: Docker Run

```bash
docker run -d \
  --name checkmk \
  -v $(pwd)/testing/custom_attrs.mk:/wato/custom_attrs.mk:ro \
  -v $(pwd)/testing/bootstrap-custom-attrs.sh:/docker-entrypoint.d/pre-start/custom_attrs.sh:ro \
  -e CMK_SITE_ID=cmk \
  -e CMK_PASSWORD=password \
  checkmk/check-mk-enterprise:2.3.0p41
```

### Option 3: Manual Copy (Running Container)

```bash
# Copy custom attributes into running container
docker cp testing/custom_attrs.mk checkmk:/omd/sites/cmk/etc/check_mk/multisite.d/wato/custom_attrs.mk

# Fix ownership
docker exec checkmk chown cmk:cmk /omd/sites/cmk/etc/check_mk/multisite.d/wato/custom_attrs.mk

# Restart CheckMK
docker exec checkmk omd restart
```

## Verifying Custom Attributes

### 1. Check File is Present

```bash
docker exec checkmk cat /omd/sites/cmk/etc/check_mk/multisite.d/wato/custom_attrs.mk
```

### 2. Test via API

```bash
# Get automation credentials
CHECKMK_URL="http://localhost:5000/cmk"
CHECKMK_USERNAME="automation"
CHECKMK_PASSWORD="your-password"

# Try to create a host with custom attribute
curl -u "${CHECKMK_USERNAME}:${CHECKMK_PASSWORD}" \
  "${CHECKMK_URL}/check_mk/api/1.0/domain-types/host_config/collections/all" \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "host_name": "test-custom-attrs",
    "folder": "/",
    "attributes": {
      "alias": "Test Host",
      "device_description": "Testing custom attributes"
    }
  }'
```

**Expected Result:**
- ✅ Success (200 OK) - custom attributes working
- ❌ Error (400 Bad Request) - custom attributes not loaded

### 3. Check in WATO UI

1. Open CheckMK UI: `http://localhost:5000/cmk/`
2. Login with `cmkadmin` or automation user
3. Setup → Hosts → Host & Service Parameters → Custom attributes
4. Verify custom attributes appear in the list

## Available Test Attributes

The `custom_attrs.mk` file defines these test attributes:

| Attribute | Type | Purpose |
|-----------|------|---------|
| `proxy_port` | TextAscii | HTTP proxy port configuration |
| `tcp_port` | TextAscii | TCP port for custom checks |
| `device_description` | TextAscii | Device description for documentation |
| `device_make` | TextAscii | Device manufacturer |
| `notes` | TextAscii | General notes field |

## Testing with Terraform

Once custom attributes are bootstrapped, you can use them in Terraform tests:

```hcl
resource "checkmk_host" "test" {
  host_name = "test-host"
  folder    = "/"

  attributes = {
    alias              = "Test Host"
    ipaddress          = "127.0.0.1"
    device_description = "Test device"  # Custom attribute
    device_make        = "Generic"      # Custom attribute
    notes              = "Test notes"   # Custom attribute
  }
}
```

## Troubleshooting

### Custom Attributes Not Working

**Problem**: API returns `Unknown attribute: 'device_description'`

**Solutions:**
1. Verify file was copied:
   ```bash
   docker exec checkmk ls -la /omd/sites/cmk/etc/check_mk/multisite.d/wato/custom_attrs.mk
   ```

2. Check file ownership:
   ```bash
   docker exec checkmk stat /omd/sites/cmk/etc/check_mk/multisite.d/wato/custom_attrs.mk
   # Should show: Uid: ( 1000/    cmk)   Gid: ( 1000/    cmk)
   ```

3. Check CheckMK logs:
   ```bash
   docker exec checkmk tail -f /omd/sites/cmk/var/log/web.log
   ```

4. Restart CheckMK:
   ```bash
   docker exec checkmk omd restart
   ```

### Bootstrap Script Not Running

**Problem**: Custom attributes file not being copied

**Solutions:**
1. Verify script is executable:
   ```bash
   docker exec checkmk ls -la /docker-entrypoint.d/pre-start/
   ```

2. Check script execution in container logs:
   ```bash
   docker logs checkmk | grep BOOTSTRAP
   ```

3. Manually run the script:
   ```bash
   docker exec checkmk /docker-entrypoint.d/pre-start/custom_attrs.sh
   ```

## Adding Your Own Custom Attributes

To add your own custom attributes for testing:

1. Edit `custom_attrs.mk`
2. Add new attribute definition:
   ```python
   {
       'add_custom_macro': True,
       'help': 'Your attribute description',
       'name': 'your_attr_name',
       'show_in_table': False,
       'title': 'Your Attribute Title',
       'topic': 'basic',  # or 'address', 'alerting', 'location', etc.
       'type': 'TextAscii'
   }
   ```
3. Restart test container
4. Verify attribute is available via API

## Production Considerations

⚠️ **These files are for TESTING ONLY**

For production:
- Define custom attributes based on your actual requirements
- Use proper change management for custom attribute creation
- Document custom attributes in your runbooks
- Consider using infrastructure-as-code for the entire CheckMK setup (not just hosts)

## References

- [CheckMK Custom Attributes Documentation](https://docs.checkmk.com/latest/en/wato_user.html#custom_attributes)
- Project docs: `../docs/custom_attributes.md`
