# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.0] - 2026-08-24

### Added

- **`long_operation_timeout`** provider setting - a separate, configurable timeout (default: 30
  minutes) for CheckMK's blocking activation and service discovery "wait for completion"
  endpoints, independent of `request_timeout`. Both endpoints redirect to themselves while the
  underlying operation is still running; they were previously bound by the same `request_timeout`
  used for ordinary fast CRUD calls (default 60s) and by Go's default 10-redirect cap, so a
  legitimately slow activation or discovery run (many sites, many services) could be cut short
  even though CheckMK was still working correctly.

### Fixed

- `strict_resource_locking = true` now actually enforces ETag validation on config activation.
  Previously the provider always sent `If-Match: *` for the activate-changes call regardless of
  this setting - the code had a comment acknowledging it wasn't implemented. It now fetches the
  current pending-changes ETag immediately before activating and sends it as `If-Match`, so a
  concurrent change to the pending change set between plan and apply surfaces as a drift error
  instead of silently activating.

### Performance

- Config activation (`activate = "auto"`, and `checkmk_activation`) no longer blocks on a fixed
  `activation_wait_time` sleep (default 5s) regardless of how long activation actually takes. It
  now polls CheckMK's activation-run completion endpoint and returns as soon as activation
  genuinely finishes, applying `activation_wait_time` only as a small propagation-margin buffer on
  top. Since `activate = "auto"` runs activation on every resource create/update/delete, this adds
  up on applies touching many resources.
- Raised the HTTP transport's idle connection limit from Go's default of 2 per host to 25 (50
  total). Terraform runs resource operations at parallelism 10 by default, all against the same
  CheckMK host; the previous default meant most of that concurrency went to connection
  setup/teardown instead of reuse.

## [1.2.0] - 2026-08-24

### Added

- **`checkmk_acknowledge`** resource - acknowledges the current problem state of a host, or of a
  single service on a host. Supports `sticky`, `persistent`, `notify`, and `expire_on`, matching
  the CheckMK REST API's acknowledgement options. Takes effect immediately and does not require
  activation.
- **`checkmk_comment`** resource - adds a comment to a host, or to a single service on a host.
  Supports `persistent` to survive a CheckMK restart. Takes effect immediately and does not
  require activation.

### Changed

- README corrected: data sources were already implemented for hosts, folders, host/service
  groups, users, contact groups, passwords, aux tags, tag groups, time periods, rules, and
  notification rules, but the README incorrectly listed "no data sources" as a limitation. Added
  a Data Sources table documenting what's available.

### Known limitations

- `checkmk_acknowledge` and `checkmk_comment` support a single host or single service only -
  `hostgroup`, `servicegroup`, and query-based targets are not yet implemented. Every attribute
  forces resource replacement on change, for the same reason as `checkmk_downtime`; delete
  removes by host/service parameters instead of CheckMK's internal id.

## [1.1.0] - 2026-08-24

### Added

- **`checkmk_service_discovery`** resource - triggers CheckMK service discovery for a host, so
  newly added or removed services are found without a manual step in the CheckMK UI. Supports all
  CheckMK discovery modes (`refresh`, `new`, `remove`, `fix_all`, `tabula_rasa`,
  `only_host_labels`, `only_service_labels`). Modes that auto-apply changes are tracked through the
  provider's existing activation handling (`activate`/`force_foreign_changes`), matching how other
  resources integrate with `checkmk_activation`.
- **`checkmk_downtime`** resource - schedules a CheckMK downtime (maintenance window) for a host,
  or for specific services on a host. Downtimes take effect immediately and do not require
  activation.

### Known limitations

- `checkmk_service_discovery`, like `checkmk_activation`, has no persistent server-side object to
  read back; it doesn't detect out-of-band drift in a host's discovered services. Re-run
  `terraform apply` to re-trigger discovery.
- `checkmk_downtime` supports the `host` and `service` downtime types only - `hostgroup`,
  `servicegroup`, and query-based (`host_by_query`/`service_by_query`) downtimes are not yet
  implemented. Every attribute forces resource replacement on change, since CheckMK's downtime
  create endpoints respond `204 No Content` with no id to target with the separate "modify
  downtime" endpoint; delete cancels by host/service parameters instead.

## [1.0.3] - 2026-08-24

First release published under this fork.

### Added

- CheckMK 2.5.x support: bumped the generated-types dependency
  ([`checkmk-api-spec`](https://github.com/BlackMesaLTD/checkmk-api-spec)) to pick up 2.5.0
  baseline types, so unlisted 2.5.x patch releases (e.g. 2.5.0p12) automatically fall back to the
  newest known 2.5 baseline for plan-time validation instead of silently dropping to hollow mode.
- CI acceptance test matrix and local docker-compose tooling extended to cover CheckMK 2.5.

### Changed

- Rebranded as a Withake-IT fork of [BlackMesaLTD/terraform-provider-checkmk](https://github.com/BlackMesaLTD/terraform-provider-checkmk):
  - Terraform provider source/address changed from `blackmesaltd/checkmk` to `withake-it/checkmk`.
  - Go module path changed to `github.com/withakedo/terraform-provider-checkmk`.
  - `goreleaser`'s `project_name` and `tfplugindocs`' `--provider-name`/`--rendered-provider-name`
    are now pinned explicitly rather than inferred from the checkout directory name.
