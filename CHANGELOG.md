# Changelog

All notable changes to GoreeCloud Home Security will be recorded here.

The project is currently in Development and has no released version.

## Unreleased

### Added

- Established the original native GoreeCloud Home Security product foundation.
- Added Platform Contract v0.2 declaration with explicit nonconformant Development status.
- Added strict local configuration loading and validation for camera registry definitions.
- Added credential-safe camera configuration using environment-variable references rather than embedded reusable secrets.
- Added loopback-only Development binding while GoreeCloud Identity and Wardveil Security authorization integration remain incomplete.
- Added sanitized read-only camera, event, health, readiness, and status APIs.
- Added bounded periodic RTSP/RTSPS probing through an optional local FFprobe adapter for unauthenticated camera streams.
- Added sanitized per-camera media state with codec, dimensions, frame rate, last-probe time, and categorical-only failure reasons.
- Added a versioned protected-worker descriptor and anonymous-pipe transport primitive so future authenticated workers can receive resolved camera credentials without placing them in command arguments or child environment variables.
- Added a fixed minimal environment for FFprobe/FFmpeg subprocesses instead of inheriting daemon environment variables.
- Added an opt-in long-running unauthenticated FFmpeg session supervisor with bounded I/O timeout, direct process execution, capped exponential restart backoff, and sanitized session state.
- Added aggregate session-running/backoff/blocked status counts to the existing bounded status API.
- Retained fail-closed credentialed FFprobe/FFmpeg execution until an authenticated worker consumes the protected descriptor contract.
- Added a shell-free FFmpeg recording-plan primitive with validated camera-scoped segment paths and bounded segment duration; no recorder process is started yet.
- Added a durable permission-restricted JSONL event journal with validation and fail-closed corruption handling.
- Added request IDs, no-store response policy, security response headers, graceful shutdown behavior, unit/race tests, and governed CI checks.
- Added product, security, privacy, recovery, architecture, benefits, branding, competitive-objective, and media-engine documentation.

### Not yet implemented

- A production authenticated RTSP media worker, exact-camera sustained-ingest evidence, offline/recovery event persistence, tamper detection, ONVIF discovery, restreaming, WebRTC/live-view delivery, recording execution, crash-safe segment indexing, retention/storage-pressure enforcement, playback, motion analysis, object detection/tracking, zones, snapshots, review UI, semantic search, face recognition, license-plate recognition, notifications, automations, and production platform-system integrations.
