# GoreeCloud Home Security — Security

## Status

Security posture: Development / incomplete. This repository is not production-approved and does not claim Wardveil Security conformance.

## Current source controls

- HTTP configuration rejects non-loopback addresses.
- RTSP/RTSPS URLs containing user-info credentials are rejected.
- Camera credential fields are environment-variable references rather than secret values.
- Camera API output omits stream URLs and secret references.
- Event storage is created with restrictive local permissions.
- Event journal corruption fails closed instead of silently dropping malformed records.
- API responses use request IDs, `no-store`, structured errors, and `nosniff`.
- HTTP server timeouts and header-size limits are configured.
- The initial Go source has no external module dependencies.

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
- Bound CPU, memory, GPU/NPU, file, descriptor, network, and queue usage.
- Avoid shell interpolation when launching media tools.
- Avoid placing reusable camera credentials in logs or externally visible command lines when a safer mechanism is available.
- Validate file paths and prevent traversal across recording/export roots.
- Validate model identity, provenance, and allowed formats before loading.
- Treat cameras and local-network devices as potentially compromised peers.

## Vulnerability reporting

Do not place secrets, private camera URLs, recordings, household information, or exploit evidence containing private data in public issues. Use the approved GoreeCloud security reporting path when it is established for this project.
