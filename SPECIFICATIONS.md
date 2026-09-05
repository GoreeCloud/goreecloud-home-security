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

This document separates current implementation from planned product direction. Planned capabilities are not implementation claims.

## 1. Role and purpose

GoreeCloud Home Security is the first-party GoreeCloud application for local camera ingest, NVR recording, video-security event detection, review, alerting, and privacy-controlled security intelligence.

It exists to provide GoreeCloud with an independently maintainable home-security platform instead of making an external complete-product NVR the permanent application layer.

## 2. Users

Target users are authorized GoreeCloud household members and administrators. The final authorization model must distinguish ordinary viewers, security reviewers, household administrators, and machine/service identities where their permissions materially differ.

## 3. Native architecture rule

The application must remain an original GoreeCloud implementation. Narrow technical foundations may include mature codecs, FFmpeg-class media tooling, ONVIF/RTSP protocol libraries, databases, inference runtimes, and hardware acceleration APIs where recreating them would increase compatibility, security, or maintenance risk. They must remain bounded dependencies, not inherited complete-product application architectures.

## 4. Current implementation — source validated locally

The current Development slice contains:

1. Strict JSON configuration loading with unknown-field rejection.
2. Camera identifiers, display names, enabled state, RTSP/RTSPS stream URLs, and environment secret references.
3. Rejection of credentials embedded in camera URLs.
4. Rejection of non-loopback HTTP exposure until the authorization boundary is implemented.
5. Sanitized public camera records containing only ID, name, and enabled state.
6. Local JSONL event journal creation with `0600` file permissions and append durability via `fsync`.
7. Fail-closed event-journal reads when stored records are malformed.
8. Read-only HTTP endpoints for health, readiness, bounded status, cameras, and events.
9. Generated request IDs, `Cache-Control: no-store`, `X-Content-Type-Options: nosniff`, and structured errors.
10. Unit tests, formatting, vet, and build validation.

These controls establish a safe foundation. They do not establish NVR functionality or production acceptance.

## 5. Required product capability domains

### 5.1 Camera and device management

Planned:

- Manual camera registration.
- ONVIF discovery and device capability inspection.
- Multiple streams per camera for live view, detection, and recording roles.
- Camera health, reconnect state, clock drift, and stream diagnostics.
- PTZ control only where explicitly authorized.
- Camera groups, locations, privacy zones, masks, and schedules.
- Secure secret references rather than stored cleartext camera credentials.

### 5.2 Media ingest and live view

Planned:

- RTSP and RTSPS ingest.
- Codec-aware stream probing.
- Managed low-latency live view and restreaming.
- Hardware decode paths where supported.
- Backpressure, reconnect, jitter, and source-failure handling.
- No default remote relay or cloud dependency.

### 5.3 Detection and tracking

Planned:

- Motion pre-filtering before expensive inference where appropriate.
- Pluggable local object-detection workers.
- CPU, GPU, NPU, and accelerator capability discovery.
- Versioned detector contracts with bounded resource use.
- Multi-object tracking and event lifecycle correlation.
- Per-camera labels, thresholds, zones, masks, dwell rules, line crossing, and object filters.
- Detection pipelines that can be disabled per camera.

### 5.4 Recording and retention

Planned:

- Continuous recording.
- Motion-based recording.
- Event/object-aware recording.
- Pre-roll and post-roll buffers.
- Retention by recording class, age, storage pressure, camera, and protected event state.
- Clip and snapshot extraction.
- Storage health and capacity safeguards.
- Deletion behavior that distinguishes active storage, indexes, thumbnails, exports, and backups.

### 5.5 Review and search

Planned:

- Unified review queue.
- Timeline with event grouping.
- Filters by camera, time, label, zone, event type, and review state.
- Fast thumbnail/preview generation.
- Portable clip/snapshot export.
- Optional semantic search using local embeddings only when explicitly enabled and privacy-reviewed.

### 5.6 Optional privacy-sensitive intelligence

Planned but **disabled by default and not implemented**:

- Face detection/recognition.
- License-plate detection/recognition.
- Semantic embeddings and natural-language video/event search.
- Person re-identification or cross-camera correlation.

Each capability requires explicit purpose, user control, retention, deletion, export, access, and evidence boundaries. External AI services must not receive private camera media without separate explicit approval.

### 5.7 Alerts and automation

Planned:

- Rule engine for object, motion, zone, dwell, line-crossing, offline-camera, tamper, and storage events.
- Rate limiting, deduplication, quiet periods, severity, and escalation.
- GoreeCloud Notify integration.
- GoreeCloud Mesh event publication with minimum necessary metadata.
- Optional MQTT/Home Assistant interoperability through bounded adapters.

## 6. API contract direction

Current v1 read-only endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/status`
- `GET /api/v1/cameras`
- `GET /api/v1/events?limit=N`

Future APIs must use explicit authentication/authorization, structured errors, request IDs, bounded pagination/filtering, rate/resource controls, timeouts, idempotency for retried mutations, and versioned deprecation rules where applicable.

## 7. Data model direction

Authoritative data domains are expected to include:

- Camera configuration and capability metadata.
- Stream-role configuration.
- Event and detection metadata.
- Recording segment indexes.
- Review state.
- Retention policy.
- Alert/rule configuration.
- Privacy-control configuration.
- Export metadata.
- Platform integration evidence/status.

Raw media and derived media must remain separately manageable from event metadata so retention and deletion can be reasoned about explicitly.

## 8. Privacy requirements

- Local-first processing is the default.
- No remote telemetry by default.
- Camera credentials must remain outside normal source and portable config.
- API/status responses must not disclose stream URLs, credentials, private media, or unnecessary device/network metadata.
- Optional biometric, plate, and semantic processing must be off by default.
- Recording and metadata retention must be explicit and configurable.
- Actual deletion and export mechanisms must precede any claim that deletion/export is supported.
- Privacy Shield runtime integration remains required and currently blocked/unaccepted.

## 9. Security requirements

- Network exposure remains loopback-only until GoreeCloud Identity and Wardveil Security integration establishes an accepted access boundary.
- Secrets must not be placed in stream URLs, logs, source, examples, or API payloads.
- Media parsing and detector workers must be isolated and resource-bounded because camera streams and media files are untrusted input.
- External command execution must avoid shell interpolation and credential leakage.
- Administrative and ordinary viewing permissions must be distinct.
- Recording/export access must be auditable without logging private media contents.
- Security review is required before any remote exposure or production deployment.

## 10. Integral Platform Systems

All seven systems are evaluated as applicable:

- GoreeCloud Manager — status/operational visibility; not integrated.
- Privacy Shield — privacy controls/evidence; required and not integrated.
- Wardveil Security — protection/security evidence; required and not integrated.
- Everkeep — backup/recovery/portability evidence; required and not integrated.
- Glaze UI — required for the future user interface; not implemented.
- GoreeCloud Mesh — registration, events, relationships, policy/coordination; not integrated.
- GoreeCloud Identity — account/session/authentication/authorization authority; not integrated.

No Stable claim is allowed while applicable required integrations remain incomplete or unverified.

## 11. Deployment direction

Primary target: self-hosted Linux systems, initially as a native service and later as a reproducible containerized deployment where appropriate. Hardware acceleration must be optional and capability-discovered rather than assumed.

The current development server is intentionally loopback-only.

## 12. Backup, recovery, and export

Required before production acceptance:

- Configuration recovery without embedding reusable camera secrets.
- Metadata/event database backup and verified restore.
- Recording policy for whether bulk video is backed up, replicated, or intentionally excluded.
- Everkeep integration for approved recovery evidence.
- Export of user-selected recordings/clips/snapshots and portable metadata where appropriate.
- Restore validation on a clean isolated target.

## 13. Testing requirements

Current tests cover configuration safety, loopback exposure policy, sanitized camera output, journal durability/permissions, malformed-journal failure, API privacy headers, and structured errors.

Future acceptance must add media fixtures, parser fuzzing, ingest/reconnect tests, detector contract tests, resource-abuse tests, retention/deletion tests, authz tests, export tests, restore tests, hardware-accelerator matrix tests, Glaze UI tests, and exact target-environment validation.

## 14. Release model

Current: Development / unreleased.

Progression to Release Candidate or Stable requires the applicable GoreeCloud release lifecycle, security/privacy/recovery evidence, platform-system acceptance, CI, exact artifact provenance, migration/rollback documentation, and target-runtime validation. Source availability or passing unit tests alone are insufficient.
