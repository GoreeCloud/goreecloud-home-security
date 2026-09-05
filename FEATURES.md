# GoreeCloud Home Security — Features

Status terms in this file are evidence boundaries, not marketing maturity claims.

| Capability | State | Evidence / boundary |
| --- | --- | --- |
| Strict JSON configuration | Implemented / locally source-validated | `internal/config` tests |
| RTSP/RTSPS camera definitions | Implemented / source-validated | Configuration only; real-camera interoperability is unverified |
| Secret-reference camera credentials | Implemented / config validation | Secrets remain outside stream URLs and normal API output |
| Loopback-only development API | Implemented / locally source-validated | Prevents ordinary remote binding before auth integration |
| Sanitized camera list | Implemented / locally source-validated | No stream URL or secret refs in API |
| RTSP/RTSPS stream probe | Implemented / locally source-validated | Optional `ffprobe`; unauthenticated streams only |
| Sanitized media health projection | Implemented / locally source-validated | State, codec, dimensions, FPS, probe time, categorical reason only |
| Protected worker descriptor | Implemented / locally source-validated | Versioned credential descriptor over anonymous pipe; secrets excluded from worker argv/environment |
| Credentialed external-media guard | Implemented / locally source-validated | FFprobe/FFmpeg execution remains blocked until protected descriptor is consumed by an authenticated media worker |
| Opt-in long-running session supervisor | Implemented / locally source-validated | Disabled by default; unauthenticated streams only; FFmpeg stream-copy to null sink with retry/backoff |
| Sanitized session status | Implemented / locally source-validated | State, attempt, timestamps and categorical reason; no URL/credentials/raw stderr |
| Minimal media-process environment | Implemented / locally source-validated | Media children receive fixed locale/timezone environment rather than daemon environment |
| Recording command-plan primitive | Implemented / locally source-validated | Shell-free FFmpeg argument plan and camera-scoped segment path; not executed |
| Append-only local event journal | Implemented / locally source-validated | JSONL, `0600`, fsync, fail-closed read |
| Health/readiness/status API | Implemented / locally source-validated | Service-level only; not production readiness |
| Event read API | Implemented / locally source-validated | Journal metadata only |
| Protected authenticated RTSP ingest worker | Planned | Descriptor contract exists; no authenticated media backend consumes it yet |
| Real-camera sustained ingest validation | Planned | No exact camera/runtime evidence yet |
| Offline/recovery event semantics | Planned | Session state exists; event transitions are not persisted yet |
| Tamper detection | Planned | Must not be inferred from ordinary connectivity failure |
| ONVIF discovery | Planned | Not implemented |
| Live view/restream | Planned | Not implemented |
| Motion detection | Planned | Not implemented |
| Object detection | Planned | Not implemented |
| Object tracking | Planned | Not implemented |
| Zones/masks/line crossing | Planned | Not implemented |
| Continuous/event recording execution | Planned | No recorder worker, segment index, or retention enforcement yet |
| Review timeline | Planned | Not implemented |
| Clips/snapshots/export | Planned | Not implemented |
| Alert rules | Planned | Not implemented |
| GoreeCloud Notify | Planned | Not integrated |
| Face recognition | Proposed / privacy-sensitive | Disabled by default; not implemented |
| License-plate recognition | Proposed / privacy-sensitive | Disabled by default; not implemented |
| Semantic search | Proposed / privacy-sensitive | Disabled by default; not implemented |
| Glaze UI | Required / blocked | No UI implementation yet |
| GoreeCloud Identity | Required / blocked | No accepted integration yet |
| Wardveil Security | Required / blocked | No accepted integration yet |
| Privacy Shield | Required / blocked | No accepted integration yet |
| Everkeep | Required / blocked | No accepted integration yet |
| GoreeCloud Mesh | Applicable / blocked | No accepted integration yet |
| GoreeCloud Manager | Applicable / blocked | No accepted integration yet |
