# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
