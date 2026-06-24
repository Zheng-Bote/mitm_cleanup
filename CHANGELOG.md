# Changelog

All notable changes to the `mitm_cleanup` component will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0] - 2026-06-24

### Added
- **extending logging**

## [v0.1.0] - 2026-06-21

### Added
- Initial release of the `mitm_cleanup` maintenance worker.
- Supports dynamically cleaning `target_fragments`, `raw_ingestion`, `system_logs`, `job_audit_logs`, `admin_audit_logs`, and `transformation_errors`.
- Configurable retention periods via JSON arguments in `os.Args[1]`.
- Native Unix Socket IPC telemetry for integration with `mitm_scheduler`.
