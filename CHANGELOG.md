# Changelog

All notable GoreeCloud Home Security repository changes are recorded here. The project is Development and has no released version.

## Unreleased

### Added

- Established the original native GoreeCloud Home Security service, strict local configuration, loopback-only Development API, sanitized camera/event/media/session status, durable JSONL event journal, Platform Contract v0.2, tests, CI, and governed project documentation.
- Added bounded periodic RTSP/RTSPS FFprobe inspection for unauthenticated definitions and a plan-only shell-free FFmpeg recording primitive.
- Added protected worker descriptor **v2** with credential transfer on inherited FD3 and sanitized status on FD4, keeping resolved camera credentials out of worker argv/environment.
- Added the GoreeCloud-owned `home-security-media-worker` RTSP/RTSPS client with DESCRIBE/SETUP/PLAY, bounded SDP/header parsing, TCP-interleaved RTP handling, and data-flow-based `media_ready` status.
- Added Digest authentication for MD5/MD5-sess/SHA-256/SHA-256-sess (`qop=auth`), RTSPS-only Basic authentication, plaintext-RTSP Basic rejection, and normal TLS certificate verification.
- Added bounded media-stall detection, sanitized categorical worker exit reasons, nonretryable auth/policy failures, retryable connectivity failures, and capped supervisor backoff.
- Added a controlled local Digest-auth RTSP integration test through first interleaved video RTP packet, plus race tests and builds for both `home-securityd` and `home-security-media-worker`.

### Not yet implemented / verified

- Exact physical-camera sustained-ingest/reconnect/stall/TLS trust evidence, persisted offline/recovery transitions, tamper detection, recorder execution/indexing, retention/storage-pressure enforcement, ONVIF, playback/live view/restreaming, motion analysis, object detection/tracking, zones/rules, review/export, alerts, privacy-sensitive intelligence, UI, and production platform-system integrations.
