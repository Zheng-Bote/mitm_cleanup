# Changelog

All notable changes to the `mitm_cleanup` component will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.3.0] - 2026-06-30

### Changed
- **Config Restructuring**: Updated database connection setup to read and parse the JSON configuration (`MITM_DB_CONFIG_JSON`) provided by the scheduler, successfully processing the nested `"db"` object format.
- **Database Connection**: The cleanup job now prioritizes the JSON configuration over direct environment variables (`MITM_DB_HOST`, etc.). Direct variables are strictly used as a fallback.
- **Audit Logging**: Added IPC audit logging during startup to explicitly record whether the database configuration was sourced from `JSON Config (MITM_DB_CONFIG_JSON)` or `Environment Variables`.

## [v0.2.0] - 2026-06-24

### Added
- **extending logging**

## [v0.1.0] - 2026-06-21

### Added
- Initial release of the `mitm_cleanup` maintenance worker.
- Supports dynamically cleaning `target_fragments`, `raw_ingestion`, `system_logs`, `job_audit_logs`, `admin_audit_logs`, and `transformation_errors`.
- Configurable retention periods via JSON arguments in `os.Args[1]`.
- Native Unix Socket IPC telemetry for integration with `mitm_scheduler`.
