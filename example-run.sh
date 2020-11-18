#!/bin/bash

export AW_RESOURCES=pods
export AW_EXPORT_TOLERATION_KEY=example-key
export AW_EXPORT_TOLERATION_OPERATOR=Exists
export AW_EXPORT_TOLERATION_EFFECT=NoSchedule
export AW_EXPORT_ALLOWED_GROUPS="trusted-users"
export AW_WASM_URI=file://$(realpath ..)/pod-toleration-policy/target/wasm32-wasi/release/pod-toleration-policy.wasm
export ADMISSION_CALLBACK_HOST=127.0.0.1

./admission-wasm
