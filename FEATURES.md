# GoreeCloud Home Security — Features

Status terms in this file are evidence boundaries, not marketing maturity claims.

| Capability | State | Evidence / boundary |
| --- | --- | --- |
| Strict JSON configuration | Implemented / locally source-validated | `internal/config` tests |
| RTSP/RTSPS camera definitions | Implemented / config-only | No live ingest yet |
| Secret-reference camera credentials | Implemented / config validation | Runtime ingest does not consume them yet |
| Loopback-only development API | Implemented / locally source-validated | Prevents ordinary remote binding before auth integration |
| Sanitized camera list | Implemented / locally source-validated | No stream URL or secret refs in API |
| Append-only local event journal | Implemented / locally source-validated | JSONL, `0600`, fsync, fail-closed read |
| Health/readiness/status API | Implemented / locally source-validated | Service-level only; not production readiness |
| Event read API | Implemented / locally source-validated | Journal metadata only |
| Live camera ingest | Planned | Not implemented |
| ONVIF discovery | Planned | Not implemented |
| Live view/restream | Planned | Not implemented |
| Motion detection | Planned | Not implemented |
| Object detection | Planned | Not implemented |
| Object tracking | Planned | Not implemented |
| Zones/masks/line crossing | Planned | Not implemented |
| Continuous/event recording | Planned | Not implemented |
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
