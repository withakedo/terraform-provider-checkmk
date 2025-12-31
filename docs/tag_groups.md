# CheckMK Tag Groups

Tag groups are used to classify and organize hosts in CheckMK. They allow you to create custom categorization schemes for your hosts beyond the built-in tags.

## Important Notes

- **No Activation Required**: Unlike hosts and folders, tag groups do NOT require activation after changes. Changes take effect immediately.
- **Built-in Tag Groups**: CheckMK has several built-in tag groups that cannot be deleted: `agent`, `piggyback`, `snmp_ds`, `address_family`, `criticality`, and `networking`.
- **ID is Immutable**: The tag group `id` cannot be changed after creation. Changing it will force resource replacement.

## Basic Example

```hcl
resource "checkmk_tag_group" "environment" {
  id    = "environment"
  title = "Environment"
  topic = "Custom Tags"
  help  = "Classify hosts by their environment"

  tags = [
    {
      id       = "prod"
      title    = "Production"
      aux_tags = []
    },
    {
      id       = "dev"
      title    = "Development"
      aux_tags = []
    },
    {
      id       = "test"
      title    = "Testing"
      aux_tags = []
    },
  ]
}
```

## Example with Auxiliary Tags

Auxiliary tags are additional tags that are automatically applied when a host is assigned a specific tag.

```hcl
resource "checkmk_tag_group" "location" {
  id    = "location"
  title = "Data Center Location"
  topic = "Infrastructure"

  tags = [
    {
      id       = "dc1"
      title    = "Data Center 1 (US-East)"
      aux_tags = ["us", "east"]
    },
    {
      id       = "dc2"
      title    = "Data Center 2 (US-West)"
      aux_tags = ["us", "west"]
    },
    {
      id       = "dc3"
      title    = "Data Center 3 (EU-Central)"
      aux_tags = ["eu", "central"]
    },
  ]
}
```

## Example with Strict ETag Locking

Enable strict resource locking to detect configuration drift:

```hcl
resource "checkmk_tag_group" "application" {
  id    = "application"
  title = "Application Type"

  # Enable strict resource locking for this specific resource
  strict_resource_locking = true

  tags = [
    {
      id       = "web"
      title    = "Web Server"
      aux_tags = []
    },
    {
      id       = "db"
      title    = "Database Server"
      aux_tags = []
    },
  ]
}
```

## Argument Reference

- `id` - (Required, Forces new resource) Unique identifier for the tag group. Must be unique across all tag groups.
- `title` - (Required) Human-readable title for the tag group.
- `topic` - (Optional) Topic for grouping tag groups in the CheckMK UI.
- `help` - (Optional) Help text describing the tag group.
- `tags` - (Required) List of tags within this tag group. Each tag has:
  - `id` - (Required) Unique identifier for the tag within this tag group.
  - `title` - (Required) Human-readable title for the tag.
  - `aux_tags` - (Optional) List of auxiliary tag IDs that are automatically applied with this tag.
- `strict_resource_locking` - (Optional) Override provider-level strict_resource_locking setting for this resource.

## Attribute Reference

No additional attributes are exported beyond those in the argument reference.

## Import

Tag groups can be imported using their ID:

```bash
terraform import checkmk_tag_group.environment environment
```

## Version Compatibility

This resource is compatible with CheckMK 2.1+ and 2.2+. The API automatically handles the differences between versions:
- CheckMK 2.2+ uses `id` field
- CheckMK 2.1 uses `ident` field (automatically mapped to `id`)
