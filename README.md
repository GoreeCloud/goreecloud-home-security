# GoreeCloud Home Security

GoreeCloud Home Security is an original, native GoreeCloud project for a local-first network video recorder (NVR), camera security, event detection, and home-security intelligence platform.

**Lifecycle:** Development  
**Release:** Unreleased development  
**Repository:** `GoreeCloud/goreecloud-home-security`  
**Canonical identifier:** `com.goreecloud.homesecurity`

This repository is not a Frigate fork. Frigate is used only as a product-capability benchmark. GoreeCloud Home Security is designed from GoreeCloud requirements with its own architecture, APIs, data model, privacy model, user experience, and platform integrations.

## Current implemented slice

The native Development foundation currently implements:

- Strict JSON configuration and credential-free RTSP/RTSPS camera definitions using environment-variable secret references.
- Loopback-only HTTP exposure while GoreeCloud Identity and Wardveil Security authorization are not integrated.
- Sanitized camera, media, session, event, health, readiness, and status APIs.
- Bounded periodic `ffprobe` inspection for enabled unauthenticated RTSP/RTSPS definitions.
- A versioned protected-worker descriptor v2: camera credentials are resolved in the parent and transferred over inherited file descriptor 3, while sanitized worker status travels over file descriptor 4. Secrets are not placed in worker argv or child environment variables.
- A GoreeCloud-owned `home-security-media-worker` RTSP/RTSPS client using only the Go standard library. It performs DESCRIBE/SETUP/PLAY, negotiates TCP-interleaved video, and reports `media_ready` only after the first negotiated video RTP packet is received.
- Digest authentication support for MD5, MD5-sess, SHA-256, and SHA-256-sess with `qop=auth`. Basic authentication is allowed only over RTSPS; Basic over plaintext RTSP fails closed.
- Strict bounded RTSP header/SDP parsing, ordinary TLS certificate verification for RTSPS, bounded read deadlines, media-stall detection, sanitized categorical failures, and capped supervisor restart backoff.
- Opt-in long-running media sessions (`media_sessions_enabled: false` by default). Credentialed sessions now use the owned protected worker instead of constructing credential-bearing FFmpeg arguments.
- A fixed minimal child environment for media processes/workers, reducing accidental secret propagation.
- A shell-free FFmpeg recording-plan primitive with camera-scoped segment paths; recording execution is still not implemented.
- A durable permission-restricted JSONL event journal and request/response hardening.
- Unit tests, a controlled local RTSP Digest-auth integration test, Go race tests, vet, build validation, and governed CI checks.

The authenticated worker is **source/test validated against a controlled local RTSP server**, not against a physical target camera. This does not establish broad camera compatibility, sustained real-camera reliability, recording, playback, detection, or production readiness.

**Recorder execution, crash-safe recording indexing, retention/storage-pressure enforcement, ONVIF discovery, playback/live restreaming, motion detection, object inference/tracking, alerting, and Glaze UI remain unimplemented.** The application is not Stable and is not production-approved.

## Development run

```bash
cp config/example.json config/local.json
# Edit config/local.json for your camera. Keep credentials out of the file.
export HOME_SECURITY_FRONT_DOOR_USER='camera-user'
export HOME_SECURITY_FRONT_DOOR_PASSWORD='camera-password'
go build -o ./home-security-media-worker ./cmd/home-security-media-worker
go run ./cmd/home-securityd -config ./config/local.json
```

Long-running sessions remain opt-in. If enabled, set `media_worker_executable` to the trusted `home-security-media-worker` path/name used for the Development run. Exact artifact provenance/path-integrity controls are still required before production acceptance.

Then, locally:

```bash
curl http://127.0.0.1:8787/healthz
curl http://127.0.0.1:8787/api/v1/status
curl http://127.0.0.1:8787/api/v1/cameras
curl 'http://127.0.0.1:8787/api/v1/events?limit=100'
```

See `SPECIFICATIONS.md`, `FEATURES.md`, `docs/ARCHITECTURE.md`, and `docs/MEDIA-ENGINE.md` for implementation boundaries and phased work.

## Validation

```bash
gofmt -w ./cmd ./internal
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/home-securityd
go build ./cmd/home-security-media-worker
```

## Privacy and security

Camera streams, recordings, snapshots, detections, credentials, biometric features, plate data, and household activity are private by default. No remote telemetry or cloud AI path is part of the current implementation. See `PRIVACY.md` and `SECURITY.md`.

## License

No project license has been selected yet. A license decision is required before a public software release is represented as open source under a specific license.
