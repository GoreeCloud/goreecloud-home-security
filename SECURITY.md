# GoreeCloud Home Security — Security

## Status

Security posture: Development / incomplete. This repository is not production-approved and does not claim Wardveil Security conformance.

## Current source controls

- HTTP configuration rejects non-loopback addresses.
- RTSP/RTSPS URLs containing user-info credentials are rejected.
- Camera credential fields are environment-variable references rather than secret values.
- Camera API output omits stream URLs and secret references.
- External FFprobe execution uses a direct argument vector and never shell interpolation.
- Media-tool stderr/raw failures are not copied into public media status; status uses bounded categorical failure reasons.
- Credentialed camera probing fails closed before an external media process starts, preventing the current implementation from placing reusable camera passwords into process command arguments.
- Parsed probe output is bounded to codec names, dimensions, and frame rate before entering public camera status.
- Recording planning validates camera IDs and segment duration and remains plan-only; `home-securityd` does not start an FFmpeg recorder yet.
- Event storage is created with restrictive local permissions.
- Event journal corruption fails closed instead of silently dropping malformed records.
- API responses use request IDs, `no-store`, structured errors, and `nosniff`.
- HTTP server timeouts and header-size limits are configured.
- The GoreeCloud-owned Go source currently has no external Go module dependencies; FFmpeg/FFprobe are optional external process foundations and are not production-pinned or accepted yet.

These are source-level controls only; they are not target-runtime or production evidence.

## Required before remote exposure

- GoreeCloud Identity authentication and authorization.
- Role/permission model for viewing live feeds, recordings, exports, camera configuration, PTZ, retention, rules, and administration.
- Wardveil Security integration and evidence.
- Rate/resource-abuse controls.
- Reverse-proxy/TLS boundary and origin policy appropriate to deployment.
- Secure session handling and CSRF protection for browser mutations.
- Audit/event model that avoids sensitive media/content logging.
- Secret-storage integration and credential rotation path.

## Media and inference threat model

Camera streams, codecs, metadata, ONVIF responses, uploaded/exported media, and detector-model inputs are untrusted. Future implementation must:

- Isolate media parsing and detector workers where practical.
- Bound CPU, memory, GPU/NPU, file, descriptor, network, process, and queue usage.
- Avoid shell interpolation when launching media tools.
- Keep reusable camera credentials out of logs and externally visible command lines; the current external-media adapter blocks credentialed cameras until a protected ingest design exists.
- Treat non-credential stream URLs as private configuration even though an authorized local host administrator may be able to observe process arguments.
- Validate file paths and prevent traversal across recording/export roots.
- Validate model identity, provenance, and allowed formats before loading.
- Treat cameras and local-network devices as potentially compromised peers.
- Pin and validate the exact FFmpeg/FFprobe package/build or container artifact before release qualification.

## Vulnerability reporting

Do not place secrets, private camera URLs, recordings, household information, or exploit evidence containing private data in public issues. Use the approved GoreeCloud security reporting path when it is established for this project.
