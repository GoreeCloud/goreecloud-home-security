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

## 1. Role and purpose

GoreeCloud Home Security is GoreeCloud's first-party local camera, NVR, event-detection, review, and home-security intelligence application. It must remain an independently maintainable GoreeCloud product rather than a Frigate fork or another complete-product NVR wrapped in GoreeCloud branding.

Target users are authorized household members and administrators. Future authorization must distinguish ordinary viewers, security reviewers, household administrators, and service identities when their permissions differ.

## 2. Native architecture and dependency rule

The application is original GoreeCloud-owned software. Narrow foundations may include mature codecs, FFmpeg/FFprobe-class media tooling, ONVIF/RTSP protocol libraries, databases, inference runtimes, and hardware-acceleration APIs when independent reimplementation would increase risk or reduce interoperability. Those components remain bounded dependencies; they do not own Home Security policy, authorization, event semantics, retention, review state, or GoreeCloud platform integration.

## 3. Current Development implementation

The current source-validated Development slice implements:

1. Strict JSON configuration with unknown-field rejection and bounded media probe/session settings.
2. Camera IDs, names, enabled state, RTSP/RTSPS URLs, and environment-variable credential references.
3. Rejection of credentials embedded in stream URLs and non-loopback API exposure.
4. Sanitized camera records that never disclose stream URLs or secret-reference names.
5. Periodic direct-process `ffprobe` inspection for enabled **unauthenticated** RTSP/RTSPS streams.
6. Bounded probe deadlines and sanitized media state: state, video/audio codec, dimensions, FPS, last-probe time, and categorical reason only.
7. A versioned protected-worker descriptor contract that resolves configured username/password references in the parent process and transfers the descriptor through an anonymous pipe intended for inherited file descriptor `3`.
8. Protected-worker arguments and environment that do not contain resolved credentials; media subprocesses receive a fixed minimal environment rather than inheriting daemon environment variables.
9. Fail-closed `credentialed_probe_blocked` behavior for current FFprobe/FFmpeg adapters. The protected descriptor is not yet consumed by an authenticated media backend.
10. An opt-in long-running FFmpeg session supervisor for **unauthenticated** streams with bounded network I/O timeout, capped exponential restart backoff, and sanitized session states. Sessions are disabled by default and stream-copy video to a null sink.
11. A shell-free FFmpeg recording-plan primitive with path-safe camera scoping and bounded segment duration. The daemon does not execute the plan.
12. A local `0600` JSONL event journal with `fsync` append durability and fail-closed malformed-record reads.
13. Read-only health, readiness, status, camera, and event APIs with request IDs, `no-store`, `nosniff`, structured errors, bounded event pagination, and aggregate sanitized media/session counts.
14. Unit tests, Go race tests, formatting, vet, build, and repository-governance validation.

This is not a working NVR release. It does not establish protected authenticated-camera compatibility, real-camera sustained-flow evidence, recording execution, retention, playback/live view, motion/object detection, or production acceptance.

## 4. Camera and media requirements

### Current

- Manual validated camera configuration.
- RTSP/RTSPS stream probing for unauthenticated streams.
- Sanitized media-health projection.
- Versioned protected credential-transfer descriptor over anonymous pipe.
- Credentialed FFprobe/FFmpeg execution blocked until an authenticated backend consumes that descriptor.
- Opt-in supervised long-running unauthenticated FFmpeg sessions with bounded I/O timeout and restart backoff.
- Sanitized session state projection.
- FFmpeg segment command planning only.

### Planned

- GoreeCloud-owned authenticated RTSP/RTSPS worker that consumes protected credentials without reconstructing a credential-bearing process command line.
- Exact real-camera sustained-flow/progress monitoring, reconnect/offline/recovery events, and clock/stream diagnostics.
- Tamper detection as a separate signal from ordinary connectivity failure.
- ONVIF discovery and device capability inspection.
- Multiple streams per camera for live view, recording, and detection roles.
- Low-latency local live view/restreaming and optional hardware decode.
- Crash-safe segment writer execution and recording indexes.
- Continuous, motion, and event/object-aware recording; pre/post-roll; clip/snapshot extraction.
- Explicit retention by age, class, camera, protected state, and storage pressure before unattended recording is enabled.

## 5. Detection, review, and automation requirements

Planned capabilities include motion/activity gating, pluggable local object detectors, CPU/GPU/NPU/accelerator discovery, versioned detector contracts, tracking, zones, masks, dwell and line-crossing rules, review queue/timeline, filters, thumbnails, portable exports, rule-based alerts, GoreeCloud Notify, Mesh events, and optional MQTT/Home Assistant interoperability through bounded adapters.

Face recognition, license-plate recognition, semantic embeddings/search, person re-identification, and cross-camera correlation are privacy-sensitive. They remain disabled by default and unimplemented until explicit purpose, access, retention, deletion, export, and evidence boundaries exist. External AI services must not receive private camera media without separate explicit approval.

## 6. API and data boundaries

Current endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/status`
- `GET /api/v1/cameras`
- `GET /api/v1/events?limit=N`

Status exposes aggregate sanitized media and session counts. Camera responses may expose bounded media and session state, but not stream URLs, credentials, secret references, raw media-tool diagnostics, or process stderr.

Session `running` indicates that the local supervised FFmpeg process started; it is not a claim that frames are continuously flowing. Current session status is transient in-memory operational state.

Future mutations and private-media APIs require GoreeCloud Identity authentication/authorization, structured errors, bounded pagination/filtering, rate/resource controls, timeouts, idempotency where retry matters, and versioned deprecation rules.

## 7. Privacy requirements

- Local-first processing and no remote telemetry by default.
- Camera credentials remain outside ordinary source and portable configuration.
- Reusable camera credentials must not be exposed in media-process command arguments, inherited child environments, logs, API payloads, or diagnostics.
- Public/status APIs must minimize device/network/media detail.
- Long-running media sessions remain disabled by default until explicitly enabled.
- Current opt-in session workers do not intentionally persist media.
- Recording/metadata retention must be explicit; deletion/export claims require actual implemented mechanisms.
- Privacy Shield runtime integration and acceptance remain required and currently blocked.

## 8. Security requirements

- API exposure remains loopback-only until GoreeCloud Identity and Wardveil Security establish an accepted access boundary.
- External media commands use direct argument vectors rather than shell interpolation.
- Media children receive a fixed minimal environment.
- Protected authenticated workers must consume credentials through the inherited descriptor channel or a stronger accepted mechanism and must not reconstruct credential-bearing process command lines.
- Camera media and metadata are untrusted inputs; parser/worker CPU, memory, GPU/NPU, process, file, descriptor, network, and queue usage must be bounded.
- Recording/export paths must prevent traversal across approved roots.
- FFmpeg/FFprobe must be pinned, reviewed, and target-runtime validated before release qualification.
- Administrative and ordinary viewing permissions must remain distinct.

## 9. Integral platform systems

All seven systems are applicable and currently incomplete/unaccepted:

- GoreeCloud Manager — operational status/visibility.
- Privacy Shield — privacy controls and evidence.
- Wardveil Security — security/protection evidence.
- Everkeep — backup, restore, recovery, portability, rollback evidence.
- Glaze UI — required user interface contract.
- GoreeCloud Mesh — registration, relationships, events, policy/coordination.
- GoreeCloud Identity — accounts, sessions, authentication, authorization, household roles, service identities.

No Stable claim is allowed while applicable required integrations remain incomplete or unverified.

## 10. Deployment, recovery, and portability

Primary target is self-hosted Linux, initially as a native service and later as reproducible containerized deployment where appropriate. Hardware acceleration is optional and capability-discovered. The current server is loopback-only, and no production FFmpeg/FFprobe package/build is selected or pinned.

Before production acceptance, Home Security requires configuration recovery without embedded camera secrets, metadata backup/restore, an explicit policy for bulk-video backup or exclusion, Everkeep integration, portable authorized media/metadata export, restore validation on a clean isolated target, and rollback behavior for the exact candidate.

## 11. Testing and release gates

Current tests cover configuration safety/defaults, loopback policy, sanitized camera/session state, media-status validation, FFprobe parsing, credentialed-process blocking, protected descriptor validation/pipe round-trip/environment minimization, session plan/process classification, supervisor lifecycle/backoff, recording-plan safety, journal durability/permissions/corruption handling, API privacy/error behavior, and Go race detection.

Future acceptance must add real camera/media fixtures, parser fuzzing, authenticated worker integration tests, sustained-flow/stall/reconnect tests, recorder/segment integrity and retention/deletion tests, detector/resource-abuse tests, authorization tests, export/restore tests, accelerator matrix tests, Glaze UI tests, exact artifact provenance, and target-environment validation.

Current release state is Development / unreleased / nonconformant. Release Candidate or Stable requires applicable lifecycle, CI, exact-candidate security/privacy/recovery evidence, platform-system acceptance, migration/rollback documentation, artifact provenance, and target-runtime validation. Source availability or passing unit tests alone are insufficient.
