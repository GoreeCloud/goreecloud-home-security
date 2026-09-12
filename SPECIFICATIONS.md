# GoreeCloud Home Security — Specifications

## Document status

- Product: GoreeCloud Home Security
- Repository: `GoreeCloud/goreecloud-home-security`
- Development model: Original native GoreeCloud application
- Lifecycle: Development
- Version: Unreleased development
- Canonical package suffix: `homesecurity`
- Canonical identifier: `com.goreecloud.homesecurity`
- Production status: Not production-approved

This document separates verified current implementation from planned product direction. Planned capabilities are not implementation claims.

## 1. Role and architecture rule

GoreeCloud Home Security is GoreeCloud's first-party local camera/NVR, event-detection, review, and home-security intelligence application. It must remain independently maintainable rather than a Frigate fork or another complete NVR wrapped in GoreeCloud branding. Narrow codec/protocol/inference/database foundations may be used, but GoreeCloud owns policy, authorization, event semantics, retention, review state, and platform integration.

## 2. Current Development implementation

The current source/test-validated slice implements strict camera/config validation, loopback-only APIs, sanitized camera/media/session state, a durable event journal, bounded FFprobe inspection, and a GoreeCloud-owned RTSP/RTSPS session worker.

The protected worker contract is schema version 2. The daemon resolves camera credential references and transfers camera ID, credential-free stream URL, credentials, and bounded worker timeout over inherited FD3. The worker receives only descriptor/status FD numbers in argv, a fixed minimal environment, and emits sanitized `media_ready` state over FD4.

The owned worker performs RTSP DESCRIBE, video-track SDP selection, SETUP with TCP interleaving, PLAY, and interleaved RTP/RTCP handling. Session `running` is published only after the first negotiated video RTP packet is received. Read deadlines provide a bounded media-stall signal.

Authentication currently supports Digest MD5, MD5-sess, SHA-256, and SHA-256-sess with `qop=auth`. Basic is allowed only over RTSPS and fails closed over plaintext RTSP. RTSPS uses normal certificate verification and TLS 1.2 or newer; there is no insecure certificate bypass.

RTSP responses are strictly bounded and malformed/folded headers fail closed. Failures are projected as categorical reasons rather than raw camera/network diagnostics. Authentication failures are nonretryable; connectivity/session-stall failures are eligible for capped supervisor backoff.

A controlled local RTSP integration test verifies a Digest challenge, authenticated DESCRIBE/SETUP/PLAY, and first interleaved RTP readiness while ensuring the password is not placed in request URIs or ordinary process surfaces. **No exact physical-camera sustained-ingest or target-runtime evidence exists yet.**

FFmpeg/FFprobe remain bounded supporting dependencies for probing and a plan-only future recording path. The long-running authenticated session path is now GoreeCloud-owned rather than a credential-bearing FFmpeg process. Recorder execution is not implemented.

## 3. Current API/data boundary

Current endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/status`
- `GET /api/v1/cameras`
- `GET /api/v1/events?limit=N`

Responses expose bounded operational state only. Stream URLs, credential references/values, worker descriptors, raw RTSP messages, and raw media-tool diagnostics are excluded.

## 4. Next camera/media requirements

1. Validate sustained flow, stall/recovery, reconnect, Digest variants, RTSPS trust behavior, and resource bounds against exact physical camera/runtime candidates.
2. Persist offline/recovery event transitions while keeping tamper detection a separate, evidence-backed signal.
3. Execute recorder workers with crash-safe segment indexing and explicit retention/storage-pressure enforcement before unattended recording.
4. Add ONVIF discovery and capability inspection.
5. Add managed local live view/restreaming and multi-stream roles.
6. Add motion gating, pluggable local detection/tracking, zones/rules, review/timeline/export, and alerts.

## 5. Privacy/security requirements

Processing remains local-first and minimized. Reusable credentials must stay out of URLs, logs, API payloads, argv, and child environment variables. Remote exposure remains prohibited until GoreeCloud Identity and Wardveil create an accepted boundary. Privacy-sensitive face/plate/semantic/re-identification/cross-camera functions remain disabled by default and unimplemented pending separate review.

Production packaging must establish trusted worker artifact/path integrity. RTSPS trust for self-signed/private-CA cameras requires an explicit trust-store mechanism rather than disabling certificate verification. Camera protocol/media input remains untrusted and requires parser fuzzing and resource/isolation tests before production acceptance.

## 6. Storage, recovery, and portability

The JSONL journal is a Development foundation, not the final metadata database. Raw media and metadata must remain separately manageable. Recorder media is not yet created. Before production acceptance, Home Security requires explicit retention/deletion policy, authorized portable export, Everkeep backup/restore design, clean-target restore tests, migration/rollback documentation, and secret-recovery boundaries.

## 7. Integral platform systems

GoreeCloud Manager, Privacy Shield, Wardveil Security, Everkeep, Glaze UI, GoreeCloud Mesh, and GoreeCloud Identity are all applicable and currently incomplete/unaccepted. No Stable or production claim is allowed while required integrations/evidence remain incomplete.

## 8. Testing and release gates

Current automated validation covers configuration, credential transfer, worker protocol, controlled Digest-auth RTSP handshake/RTP readiness, Basic-over-RTSP rejection, SDP validation, supervisor retry policy, race detection, API minimization, journal durability, and buildability of both binaries.

Future acceptance requires exact physical-camera/runtime tests, sustained-flow/reconnect/stall matrices, parser fuzzing, TLS/private-CA behavior, process/resource isolation, recorder integrity/retention/deletion, authorization, export/restore, accelerator tests, Glaze UI tests, exact artifact provenance, and platform-system evidence.

Current release state remains Development / unreleased / nonconformant.
