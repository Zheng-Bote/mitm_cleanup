#!/usr/bin/sh


cargo build --release --target x86_64-unknown-linux-musl
cp target/x86_64-unknown-linux-musl/release/mitm-cleanup /home/zb_bamboo/DEV/__NEW__/Go/mitm-2/scheduler/mitm_scheduler/bin/.

