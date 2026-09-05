# GoreeCloud Home Security — Features

Status terms in this file are evidence boundaries, not marketing maturity claims.

| Capability | State | Evidence / boundary |
| --- | --- | --- |
| Strict JSON configuration | Implemented / locally source-validated | `internal/config` tests |
| RTSP/RTSPS camera definitions | Implemented / source-validated | Configuration only; continuous ingest is not implemented |
| Secret-reference camera credentials | Implemented / config validation | External media probing deliberately refuses credentialed cameras pending protected ingest |
| Loopback-only development API | Implemented / locally source-validated | Prevents ordinary remote binding before auth integration |
| Sanitized camera list | Implemented / locally source-validated | No stream URL or secret refs in API |
| RTSP/RTSPS stream probe | Implemented / locally source-validated | Optional `ffprobe`; unauthenticated streams only; no sustained ingest claim |
| Sanitized media health projection | Implemented / locally source-validated | State, codec, dimensions, FPS, probe time, categorical reason only |
| Credentialed media-process guard | Implemented / locally source-validated | Fails closed before external process execution to avoid password argv exposure |
| Recording command-plan primitive | Implemented / locally source-validated | Shell-free FFmpeg argument plan and camera-scoped segment path; not executed |
| Append-only local event journal | Implemented / locally source-validated | JSONL, `0600`, fsync, fail-closed read |
| Health/readiness/status API | Implemented / locally source-validated | Service-level only; not production readiness |
| Event read API | Implemented / locally source-validated | Journal metadata only |
| Long-running live camera ingest | Planned | Not implemented |
| Protected authenticated RTSP ingest | Planned | Current external-media path blocks credentialed cameras |
| ONVIF discovery | Planned | Not implemented |
| Live view/restream | Planned | Not implemented |
| Motion detection | Planned | Not implemented |
| Object detection | Planned | Not implemented |
| Object tracking | Planned | Not implemented |
| Zones/masks/line crossing | Planned | Not implemented |
| Continuous/event recording execution | Planned | No recorder worker or retention enforcement yet |
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
