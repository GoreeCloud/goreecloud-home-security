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
- Added a durable permission-restricted JSONL event journal with validation and fail-closed corruption handling.
- Added request IDs, no-store response policy, security response headers, and graceful shutdown behavior.
- Added unit tests and CI checks for formatting, vetting, tests, buildability, and governed repository documentation.
- Added product, security, privacy, recovery, architecture, benefits, branding, and competitive-objective documentation.

### Not yet implemented

- Live RTSP/RTSPS ingestion, restreaming, WebRTC/live-view delivery, recording, playback, retention enforcement, motion analysis, object detection, tracking, zones, snapshots, review UI, semantic search, face recognition, license-plate recognition, notifications, automations, and production platform-system integrations.
