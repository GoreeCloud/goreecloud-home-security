# GoreeCloud Home Security — Security

## Status

Security posture: Development / incomplete. This repository is not production-approved and does not claim Wardveil Security conformance.

## Current source controls

- HTTP configuration rejects non-loopback addresses.
- RTSP/RTSPS URLs containing user-info credentials are rejected.
- Camera credential fields are environment-variable references rather than secret values.
- Camera API output omits stream URLs and secret references.
- External FFprobe/FFmpeg execution uses direct argument vectors and never shell interpolation.
- Media subprocesses receive a fixed minimal environment (`LANG`, `LC_ALL`, `TZ`) rather than inheriting daemon environment variables, preventing configured camera secret variables from being propagated to those children.
- Media-tool stderr/raw failures are discarded instead of being copied into public media/session status; status uses bounded categorical reasons.
- Credentialed FFprobe/FFmpeg execution still fails closed before process execution.
- A versioned protected-worker descriptor can resolve credential references in the parent process and serialize username/password material through an anonymous pipe intended for an inherited file descriptor. The protected worker surface does not put these values into command arguments or child environment variables.
- No authenticated media worker consumes that descriptor yet, so this is a protected transfer contract—not a claim of working authenticated RTSP ingest.
- Long-running FFmpeg session supervision is opt-in, unauthenticated-only, uses bounded network I/O timeout, and applies capped exponential restart backoff.
- Parsed probe output and public session state are bounded before entering API status.
- Recording planning validates camera IDs and segment duration and remains plan-only; `home-securityd` does not start an FFmpeg recorder yet.
- Event storage is created with restrictive local permissions and corruption fails closed.
- API responses use request IDs, `no-store`, structured errors, and `nosniff`; server timeouts and header-size limits are configured.
- The GoreeCloud-owned Go source has no external Go module dependencies; FFmpeg/FFprobe remain optional external process foundations and are not production-pinned or accepted yet.

These are source-level controls only; they are not target-runtime or production evidence.

## Protected credential-transfer boundary

The descriptor contract uses an anonymous pipe so reusable camera credentials do not need to appear in `/proc/.../cmdline` or the child environment. The descriptor is size-bounded, versioned, strict-decoded, validates credential-free RTSP/RTSPS URLs, and requires username/password values as a pair.

This reduces accidental exposure; it does not protect credentials from a privileged host administrator, kernel compromise, process-memory inspection, or an unsafe future worker implementation. A production authenticated worker must consume the inherited descriptor directly and must not reconstruct a credential-bearing FFmpeg command line.

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

Camera streams, codecs, metadata, ONVIF responses, uploaded/exported media, and detector-model inputs are untrusted. Future implementation must isolate/bound media and detector workers, keep secrets out of logs and externally visible arguments, validate paths and model provenance, treat network cameras as potentially compromised peers, and pin/validate the exact FFmpeg/FFprobe package/build before release qualification.

A non-credential stream URL can still be visible in the opt-in FFmpeg process command line. It remains private configuration and the current Development implementation assumes the local host administrator is trusted. Credential-bearing URLs remain prohibited.

## Vulnerability reporting

Do not place secrets, private camera URLs, recordings, household information, or exploit evidence containing private data in public issues. Use the approved GoreeCloud security reporting path when it is established for this project.
