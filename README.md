# Terraform Provider for CheckMK

A Terraform provider for managing [CheckMK](https://checkmk.com/) monitoring infrastructure as code using the CheckMK REST API.

Maintained by [Withake-IT](https://withake-it.de).

> **Fork notice:** This project is a fork of [BlackMesaLTD/terraform-provider-checkmk](https://github.com/BlackMesaLTD/terraform-provider-checkmk),
> licensed under the [Mozilla Public License 2.0](LICENSE). It continues development independently under
> [withakedo/terraform-provider-checkmk](https://github.com/withakedo/terraform-provider-checkmk), including
> ongoing compatibility updates for newer CheckMK releases. All credit for the original implementation goes to
> the upstream authors.

## Features

- **Full CRUD operations** for hosts, folders, users, rules, and more
- **Version-aware API handling** - automatically adapts to CheckMK 2.2.x, 2.3.x, 2.4.x, and 2.5.x differences
- **Plan-time validation** - field names and enum values validated during `terraform plan` using version-specific types
- **OpenAPI-driven type safety** - field validation based on actual CheckMK API schemas
- **Optimistic locking** - ETag-based concurrency control prevents conflicting updates
- **Flexible activation** - auto-activate changes or batch them with explicit activation resources
- **Import support** - bring existing CheckMK configuration under Terraform management
- **Organized documentation** - resources grouped by category (Hosts & Folders, Rules, Users, etc.)

## Related Projects

This provider uses generated types from the companion repository:

| Repository | Purpose |
|------------|---------|
| [checkmk-api-spec](https://github.com/BlackMesaLTD/checkmk-api-spec) | OpenAPI specs and generated Go types for CheckMK REST API |

The `checkmk-api-spec` repository provides:
- **42 baseline packages** covering all API variations across CheckMK versions
- **Runtime version mapping** - any CheckMK version maps to correct types
- **Field validators** - enum values, required fields, type information
- **Union descriptions** - merged field documentation with version annotations

## Requirements

- **Terraform**: >= 1.0
- **CheckMK**: 2.2.x, 2.3.x, 2.4.x, or 2.5.x with REST API enabled
- **Go**: >= 1.21 (for development only)

## Installation

```hcl
terraform {
  required_providers {
    checkmk = {
      source  = "withake-it/checkmk"
      version = "~> 0.1"
    }
  }
}
```

## Authentication

Create an automation user in CheckMK:

1. Navigate to **Setup → Users → Users**
2. Click **Add user**
3. Select **Automation user (for scripts, API access)**
4. Note the **username** and **automation secret**

## Provider Configuration

```hcl
provider "checkmk" {
  url      = "https://monitoring.example.com/mysite"
  username = "automation"
  password = var.checkmk_password

  # Optional settings
  activate              = "auto"    # "auto" or "manual" (default: "manual")
  strict_resource_locking = true    # Use ETags for optimistic locking
  request_timeout       = 60        # API timeout in seconds
  max_retries           = 3         # Retry count for transient failures
  insecure_skip_verify  = false     # Skip TLS verification (not recommended)
}
```

Environment variables are also supported:
- `CHECKMK_URL`
- `CHECKMK_USERNAME`
- `CHECKMK_PASSWORD`

## Resources

### Configuration

| Resource | Description |
|----------|-------------|
| `checkmk_folder` | Hierarchical folder structure for organizing hosts |
| `checkmk_tag_group` | Host tag groups with predefined choices |
| `checkmk_aux_tag` | Auxiliary tags for tag group dependencies |
| `checkmk_time_period` | Time period definitions for rules and notifications |

### Hosts & Groups

| Resource | Description |
|----------|-------------|
| `checkmk_host` | Monitored hosts with attributes, tags, and labels |
| `checkmk_host_group` | Logical groupings of hosts |
| `checkmk_service_group` | Logical groupings of services |

### Users & Access

| Resource | Description |
|----------|-------------|
| `checkmk_user` | User accounts with roles and contact settings |
| `checkmk_contact_group` | Contact groups for notifications |
| `checkmk_password` | Stored credentials for integrations |

### Rules

| Resource | Description |
|----------|-------------|
| `checkmk_rule` | Generic rule resource for any CheckMK ruleset |
| `checkmk_notification_rule` | Notification routing and filtering rules |
| `checkmk_host_labels` | Host label rules |
| `checkmk_service_labels` | Service label rules |

### Rule Wrappers (Typed Resources)

Simplified interfaces for common rule types:

| Resource | Description |
|----------|-------------|
| `checkmk_host_check_interval` | Host check interval configuration |
| `checkmk_service_check_interval` | Service check interval configuration |
| `checkmk_host_notification_period` | Host notification time periods |
| `checkmk_service_notification_period` | Service notification time periods |
| `checkmk_host_max_check_attempts` | Host check retry configuration |
| `checkmk_service_max_check_attempts` | Service check retry configuration |
| `checkmk_host_retry_interval` | Host retry interval configuration |
| `checkmk_service_retry_interval` | Service retry interval configuration |
| `checkmk_host_check_commands` | Host check command configuration |
| `checkmk_service_custom_rule` | Custom service rules |

### Operations

| Resource | Description |
|----------|-------------|
| `checkmk_activation` | Explicit activation checkpoint for staged changes |
| `checkmk_service_discovery` | Triggers service discovery for a host |
| `checkmk_downtime` | Schedules a host or service maintenance window |
| `checkmk_acknowledge` | Acknowledges the current problem state of a host or service |
| `checkmk_comment` | Adds a comment to a host or service |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| `checkmk_host` | Look up an existing host and its attributes |
| `checkmk_folder` | Look up an existing folder |
| `checkmk_host_group` | Look up an existing host group |
| `checkmk_service_group` | Look up an existing service group |
| `checkmk_user` | Look up an existing user account |
| `checkmk_contact_group` | Look up an existing contact group |
| `checkmk_password` | Look up an existing stored credential |
| `checkmk_aux_tag` | Look up an existing auxiliary tag |
| `checkmk_tag_group` | Look up an existing host tag group |
| `checkmk_time_period` | Look up an existing time period |
| `checkmk_rule` | Look up an existing rule by ID |
| `checkmk_notification_rule` | Look up an existing notification rule by ID |

## Quick Start

### Create a Folder and Host

```hcl
resource "checkmk_folder" "production" {
  path  = "/Production"
  title = "Production Systems"
}

resource "checkmk_host" "web_server" {
  name   = "web-01.example.com"
  folder = checkmk_folder.production.path

  ipaddress = "10.1.2.3"
  alias     = "Production Web Server"

  labels = {
    environment = "production"
    team        = "platform"
  }
}
```

### Create a Rule with Conditions

```hcl
resource "checkmk_service_check_interval" "critical_services" {
  interval    = 60
  description = "1-minute check interval for critical services"

  conditions {
    host_labels = {
      environment = "production"
    }
    service_labels = {
      critical = "true"
    }
  }
}
```

### Manage Activation Explicitly

```hcl
provider "checkmk" {
  # ... credentials ...
  activate = "manual"  # Don't auto-activate
}

resource "checkmk_host" "servers" {
  for_each = var.servers
  # ... host configuration ...
}

# Activate all changes at once
resource "checkmk_activation" "deploy" {
  depends_on = [checkmk_host.servers]

  # Re-activate when hosts change
  triggers = {
    hosts = md5(jsonencode(checkmk_host.servers))
  }
}
```

## Version Compatibility

| CheckMK Version | Support Status |
|-----------------|----------------|
| 2.5.x | Fully supported |
| 2.4.x | Fully supported |
| 2.3.x | Fully supported |
| 2.2.x | Fully supported |

The provider automatically detects the CheckMK version and adjusts API calls accordingly. Type
validation relies on generated types from
[checkmk-api-spec](https://github.com/BlackMesaLTD/checkmk-api-spec), which ships baseline types
for specific patch releases per minor version; unlisted patch releases automatically fall back to
the newest known baseline for that minor version. If you connect to a CheckMK version newer than
any known baseline (e.g. a brand new minor release), the provider falls back to `type_mode =
"hollow"` with a warning instead of failing - resources are still fully functional, just without
plan-time field/enum validation until the companion spec repo (and this provider's dependency on
it) is updated.

## Current Limitations

### API Coverage
- **Partial resource coverage** - Not all CheckMK REST API endpoints are implemented yet. Priority has been given to hosts, folders, users, and rules.
- **Partial data source coverage** - Data sources exist for the most commonly-referenced object types (hosts, folders, groups, users, rules, etc. - see the table above), but not every resource type has a matching data source yet.

### Validation
- **Plan-time validation** - Enum values of known fields (like `tag_agent`) are validated during `terraform plan` when connected to CheckMK. Invalid values produce errors before apply.
- **Custom attributes** - The `attributes` map is open: user-defined custom attributes are accepted without a prefix and passed straight through to the API, which is the source of truth for whether they exist. Keys that aren't built-in fields are validated by the API at apply time.
- **Built-in tag groups** - Built-in host tag groups may be written with or without the `tag_` prefix (e.g. `agent` or `tag_agent`); the provider promotes the unprefixed form to the API automatically and keeps your configured key in state. Custom tag groups should be written with the explicit `tag_` prefix.
- **Hollow mode** - Set `type_mode = "hollow"` to skip plan-time validation entirely (useful for testing or when version types are unavailable).

### Rules
- **Complex rule values** - Some rule value types with deeply nested structures may require JSON encoding in the `value` attribute.
- **Rule ordering** - Rule order within a ruleset is managed but bulk reordering operations are not atomic.

### Operations
- **Single-site only** - Distributed monitoring with multiple sites is not yet supported.
- **No bulk operations** - Each resource is created/updated individually; bulk host creation is not implemented.
- **Activation scope** - Activation applies to all pending changes, not just Terraform-managed resources.
- **Service discovery is not re-readable** - `checkmk_service_discovery` triggers a discovery run and records its
  result, but (like `checkmk_activation`) has no persistent server-side object to read back; drift in a host's
  discovered services is not detected automatically. Re-run `terraform apply` to re-trigger discovery.
- **Downtimes: host and service only, no in-place modify** - `checkmk_downtime` supports `host` and `service`
  downtime types; `hostgroup`, `servicegroup`, and query-based downtimes are not yet implemented. Every attribute
  forces replacement on change, since CheckMK's downtime create endpoints don't return a lookup-friendly id to
  target with the separate "modify downtime" endpoint. Deleting the resource cancels the downtime by host/service
  parameters rather than by CheckMK's internal downtime id.
- **Acknowledgements and comments: host and single service only, no in-place modify** - `checkmk_acknowledge` and
  `checkmk_comment` support a single host or a single service on a host; `hostgroup`, `servicegroup`, and
  query-based targets are not yet implemented. Every attribute forces replacement on change, for the same reason
  as `checkmk_downtime` - CheckMK's create endpoints don't return a lookup-friendly id. Deleting either resource
  removes the acknowledgement/comment by host/service parameters rather than by CheckMK's internal id.

### Version Differences
- **Schema variations** - Some attributes are only available in certain CheckMK versions. The provider accepts all attributes but the API may reject version-incompatible values.
- **Enum value changes** - Valid enum values (like `tag_agent` choices) vary between versions. The union descriptions document these differences.

## Development

### Building

```bash
go build -o terraform-provider-checkmk
```

### Testing

```bash
# Unit tests
go test ./...

# Acceptance tests (requires running CheckMK instance)
TF_ACC=1 \
CHECKMK_URL="http://localhost:5000/mysite" \
CHECKMK_USERNAME="automation" \
CHECKMK_PASSWORD="secret" \
go test -v ./internal/provider -timeout 30m
```

### Local Development

```bash
# Add to ~/.terraformrc
cat >> ~/.terraformrc <<EOF
provider_installation {
  dev_overrides {
    "withake-it/checkmk" = "/path/to/terraform-provider-checkmk"
  }
  direct {}
}
EOF
```

## Documentation

- [CHANGELOG](CHANGELOG.md) - notable changes to this provider
- [CheckMK REST API Documentation](https://docs.checkmk.com/latest/en/rest_api.html)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [checkmk-api-spec Repository](https://github.com/BlackMesaLTD/checkmk-api-spec) - Generated types and version mappings

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.
