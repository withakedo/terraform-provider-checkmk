# Terraform Provider for CheckMK

A Terraform provider for managing [CheckMK](https://checkmk.com/) monitoring infrastructure as code using the CheckMK REST API.

## Requirements

- **Terraform**: >= 1.0
- **CheckMK**: 2.2.x, 2.3.x, or 2.4.x with REST API enabled
- **Go**: >= 1.21 (for development only)

## Installation

```hcl
terraform {
  required_providers {
    checkmk = {
      source  = "blackmesaltd/checkmk"
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
| 2.4.x | Fully supported |
| 2.3.x | Fully supported |
| 2.2.x | Fully supported |

The provider automatically detects the CheckMK version and adjusts API calls accordingly.

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
    "blackmesaltd/checkmk" = "/path/to/terraform-provider-checkmk"
  }
  direct {}
}
EOF
```

## Documentation

- [CheckMK REST API Documentation](https://docs.checkmk.com/latest/en/rest_api.html)
- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.
