# GoreeCloud Home Security — Security

## Status

Security posture: Development / incomplete. This repository is not production-approved and does not claim Wardveil Security conformance.

## Current controls

- Development HTTP configuration rejects non-loopback addresses.
- Camera RTSP/RTSPS URLs containing user-info credentials are rejected; credential fields are environment-variable references.
- Public camera/media/session API output omits stream URLs, secret references, credentials, raw stderr, and raw network diagnostics.
- The protected worker descriptor is versioned and strict-decoded. The parent resolves camera secrets, sends the descriptor over inherited FD3, and receives sanitized status only on FD4. Worker argv contains only `--descriptor-fd=3 --status-fd=4`; worker child environment is fixed to `LANG`, `LC_ALL`, and `TZ`.
- The GoreeCloud-owned RTSP worker consumes credentials directly rather than reconstructing a credential-bearing FFmpeg command line.
- Digest authentication supports MD5/MD5-sess/SHA-256/SHA-256-sess with qop=auth. Basic authentication is rejected over plaintext RTSP and accepted only when protected by RTSPS.
- RTSPS uses ordinary Go TLS certificate verification with a TLS 1.2 minimum; certificate verification is not disabled. Cameras using untrusted/self-signed certificates will require an explicit future trust-store design rather than an insecure bypass.
- RTSP parsing is bounded: line/header counts and sizes, response bodies/SDP, URLs, auth challenges, and transport/control values are constrained. Folded headers are rejected.
- Session readiness is data-flow based: `media_ready` is emitted only after a negotiated video RTP packet arrives on the interleaved channel.
- Network/media failures collapse to bounded categorical reasons. Authentication failures are nonretryable in the supervisor to avoid indefinite credential retry loops; connectivity/stall failures use capped backoff.
- Media/tool subprocesses use direct argument vectors, no shell interpolation, fixed minimal environments, bounded timeouts, and discarded stdout/stderr where applicable.
- Event journal permissions/durability and API request metadata hardening remain in place.

These are source/test controls, not target-runtime security acceptance. The controlled RTSP test server does not substitute for hostile-camera testing, parser fuzzing, kernel/process isolation, package provenance, or Wardveil review.

## Remaining security gates

Before remote or production exposure the project still requires GoreeCloud Identity authentication/authorization, Wardveil integration, role separation, rate/resource-abuse controls, trusted worker artifact/path integrity, secret storage/rotation, production TLS/reverse-proxy policy, audit design, parser fuzzing, and target-camera/network testing.

The worker executable setting currently validates that the configured basename is `home-security-media-worker`; production packaging must additionally establish exact artifact provenance, ownership/permissions, immutable/trusted execution path, and update integrity.

## Media threat model

Camera streams, RTSP headers/SDP, codecs, ONVIF responses, media files, and detector inputs are untrusted. Future work must isolate/bound parser and detector workers, enforce CPU/memory/file/descriptor/network limits, validate recording/export paths, pin media dependencies, and treat network cameras as potentially compromised peers.

## Vulnerability reporting

Do not place secrets, private camera URLs, recordings, household information, or exploit evidence containing private data in public issues. Use the approved GoreeCloud security reporting path when established for this project.
