# MitM Cleanup Job (Rust)

Linux-targeted Rust replacement for `mitm_cleanup`. It preserves the scheduler Unix-socket protocol, database configuration contract, retention arguments, batched deletion, `VACUUM ANALYZE`, timeout, and progress/audit reporting.

## Build and verify on Linux

```bash
cargo build --release
cargo test
install -Dm755 target/release/mitm-cleanup bin/mitm-cleanup
```

## Run

```bash
./bin/mitm-cleanup '{"target_fragments_retention_days":7,"raw_ingestion_orphan_days":14,"timeout_minutes":60}'
```

The executable obtains database credentials from `MITM_DB_CONFIG_JSON` or the `MITM_DB_HOST`, `MITM_DB_PORT`, `MITM_DB_USER`, `MITM_DB_PASSWORD`, and `MITM_DB_NAME` variables. `MITM_DB_SSLMODE=true` enables required TLS. When `RUN_ID` and `SCHEDULER_SOCKET_PATH` are supplied, the existing scheduler credential and telemetry protocol is used.

All table names and predicates are compile-time constants. Retention values are SQL parameters and negative retention periods are rejected before a deletion query is executed.
