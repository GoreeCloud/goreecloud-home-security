# GoreeCloud Home Security — Features

Status terms are evidence boundaries, not marketing maturity claims.

| Capability | State | Evidence / boundary |
| --- | --- | --- |
| Strict JSON configuration | Implemented / source-tested | `internal/config` tests |
| Credential-free RTSP/RTSPS camera definitions | Implemented / source-tested | Credentials use secret references, not URLs |
| Loopback-only Development API | Implemented / source-tested | Remote exposure remains blocked pending Identity/Wardveil |
| Sanitized camera/media/session status | Implemented / source-tested | No URL, credentials, raw stderr, or raw network diagnostics |
| Periodic FFprobe inspection | Implemented / source-tested | Unauthenticated definitions only |
| Protected worker descriptor v2 | Implemented / source-tested | Secret descriptor FD3; sanitized status FD4; no secrets in argv/environment |
| GoreeCloud-owned authenticated RTSP worker | Implemented / controlled integration-tested | DESCRIBE/SETUP/PLAY + TCP interleaved RTP readiness against local test server |
| RTSP Digest authentication | Implemented / controlled integration-tested | MD5/MD5-sess/SHA-256/SHA-256-sess, qop=auth |
| RTSP Basic authentication policy | Implemented / source-tested | Allowed only over RTSPS; plaintext RTSP Basic fails closed |
| RTSPS transport | Implemented / source-tested | Normal TLS certificate verification; no insecure skip-verify path |
| Bounded RTSP parser and SDP selection | Implemented / source-tested | Bounded lines/headers/body/control URLs; first video control selected |
| Media-flow readiness | Implemented / controlled integration-tested | Session becomes running only after first negotiated video RTP packet |
| Media stall detection | Implemented / source-tested | Bounded read deadline; categorized as `session_stalled` |
| Session supervision/backoff | Implemented / source-tested | Opt-in; auth failures nonretryable, connectivity failures retryable |
| Minimal worker/process environment | Implemented / source-tested | Fixed locale/timezone only |
| Recording command-plan primitive | Implemented / source-tested | Not executed; no working recorder claim |
| Append-only event journal | Implemented / source-tested | JSONL, `0600`, fsync, fail-closed reads |
| Exact real-camera sustained-ingest validation | Planned | No physical camera/runtime evidence yet |
| Persisted offline/recovery events | Planned | Runtime state exists; transitions are not journaled yet |
| Tamper detection | Planned | Must remain distinct from ordinary connectivity failures |
| Recorder execution and crash-safe segment index | Planned | Not implemented |
| Retention/storage-pressure enforcement | Planned | Required before unattended recording |
| ONVIF discovery | Planned | Not implemented |
| Live view/restream/WebRTC | Planned | Not implemented |
| Motion detection | Planned | Not implemented |
| Object detection/tracking | Planned | Not implemented |
| Zones/masks/line crossing | Planned | Not implemented |
| Review timeline/clips/snapshots/export | Planned | Not implemented |
| Alerts / GoreeCloud Notify | Planned | Not integrated |
| Face recognition | Proposed / privacy-sensitive | Disabled by default; not implemented |
| License-plate recognition | Proposed / privacy-sensitive | Disabled by default; not implemented |
| Semantic search / re-identification | Proposed / privacy-sensitive | Disabled by default; not implemented |
| Glaze UI | Required / blocked | No UI implementation yet |
| GoreeCloud Identity | Required / blocked | No accepted integration yet |
| Wardveil Security | Required / blocked | No accepted integration yet |
| Privacy Shield | Required / blocked | No accepted integration yet |
| Everkeep | Required / blocked | No accepted integration yet |
| GoreeCloud Mesh | Applicable / blocked | No accepted integration yet |
| GoreeCloud Manager | Applicable / blocked | No accepted integration yet |
