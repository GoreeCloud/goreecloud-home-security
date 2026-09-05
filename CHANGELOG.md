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
- Added fail-closed blocking for credentialed external media probing so reusable camera passwords are not placed into subprocess command arguments.
- Added a shell-free FFmpeg recording-plan primitive with validated camera-scoped segment paths and bounded segment duration; no recorder process is started yet.
- Added a durable permission-restricted JSONL event journal with validation and fail-closed corruption handling.
- Added request IDs, no-store response policy, security response headers, and graceful shutdown behavior.
- Added unit tests and CI checks for formatting, vetting, tests, buildability, and governed repository documentation.
- Added product, security, privacy, recovery, architecture, benefits, branding, competitive-objective, and media-engine documentation.

### Not yet implemented

- Long-running RTSP/RTSPS ingest, protected authenticated media workers, ONVIF discovery, restreaming, WebRTC/live-view delivery, recording execution, playback, retention enforcement, motion analysis, object detection, tracking, zones, snapshots, review UI, semantic search, face recognition, license-plate recognition, notifications, automations, and production platform-system integrations.
