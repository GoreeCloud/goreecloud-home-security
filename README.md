# GoreeCloud Home Security

GoreeCloud Home Security is an original, native GoreeCloud project for a local-first network video recorder (NVR), camera security, event detection, and home-security intelligence platform.

**Lifecycle:** Development  
**Release:** Unreleased development  
**Repository:** `GoreeCloud/goreecloud-home-security`  
**Canonical identifier:** `com.goreecloud.homesecurity`

This repository is not a Frigate fork. Frigate is used only as a product-capability benchmark. GoreeCloud Home Security is being designed from GoreeCloud requirements with its own architecture, APIs, data model, privacy model, user experience, and platform integrations.

## Current implemented slice

The native Development foundation currently implements:

- Strict JSON configuration loading and validation.
- RTSP/RTSPS camera definitions with credentials prohibited from stream URLs.
- Environment-variable references for camera credentials.
- Loopback-only HTTP binding while GoreeCloud Identity and Wardveil Security network authorization are not integrated.
- Sanitized camera inventory APIs that do not expose stream URLs or secret references.
- Periodic bounded RTSP/RTSPS stream probing through an optional local `ffprobe` adapter for **unauthenticated** stream definitions.
- Sanitized per-camera media state with codec, resolution, frame rate, last-probe timestamp, and categorical failure reason only.
- Fail-closed blocking of credentialed external-media probing so reusable camera passwords are not placed into subprocess command lines.
- A shell-free FFmpeg recording-plan primitive with camera-scoped segment paths; the plan is not executed yet.
- A local append-only JSONL event journal with restrictive file permissions and fail-closed reads.
- `GET /healthz`, `GET /readyz`, `GET /api/v1/status`, `GET /api/v1/cameras`, and `GET /api/v1/events`.
- Request IDs, no-store response policy, structured API errors, and bounded event pagination.
- Unit tests, `go vet`, build validation, and repository-governance checks.

This is still a Development media/control-plane foundation. **Continuous ingest, authenticated RTSP media workers, recording execution, retention enforcement, playback, live restreaming, motion detection, object inference/tracking, alerting, and the Glaze UI are not implemented yet.** The application is not Stable and is not production-approved.

## Development run

```bash
cp config/example.json config/local.json
# Edit config/local.json for your camera. Keep credentials out of the file.
export HOME_SECURITY_FRONT_DOOR_USER='camera-user'
export HOME_SECURITY_FRONT_DOOR_PASSWORD='camera-password'
go run ./cmd/home-securityd -config ./config/local.json
```

Then, locally:

```bash
curl http://127.0.0.1:8787/healthz
curl http://127.0.0.1:8787/api/v1/status
curl http://127.0.0.1:8787/api/v1/cameras
curl 'http://127.0.0.1:8787/api/v1/events?limit=100'
```

If `ffprobe` is installed and a camera definition does not require credentials, Home Security will probe that stream on the configured interval. Credentialed cameras currently report `credentialed_probe_blocked`; this is a deliberate protection boundary, not an authentication failure. See `docs/MEDIA-ENGINE.md`.

The development server intentionally rejects non-loopback listen addresses until an accepted authentication, authorization, and network-exposure boundary exists.

## Target capability direction

The planned product is a capability superset for GoreeCloud home-security use, including:

- ONVIF discovery and standards-based camera onboarding.
- RTSP/RTSPS ingest and managed live restreaming.
- Motion pre-filtering to reduce unnecessary inference work.
- Pluggable local object detection across CPU, GPU, NPU, and accelerator backends.
- Multi-object tracking, zones, masks, dwell/crossing rules, and event correlation.
- Continuous, motion, event, and object-aware recording policies.
- Timeline, review queue, clips, snapshots, playback, and export.
- Local alert rules and GoreeCloud Notify integration.
- Optional local face recognition, license-plate recognition, and semantic search behind explicit privacy controls.
- GoreeCloud Identity authorization, Wardveil Security protection, Privacy Shield controls, Everkeep backup/recovery, GoreeCloud Mesh coordination, Manager status, and Glaze UI.
- Home Assistant and MQTT interoperability where it provides value without becoming the core architecture.

See `SPECIFICATIONS.md`, `FEATURES.md`, `docs/ARCHITECTURE.md`, and `docs/MEDIA-ENGINE.md` for the implementation boundary and phased roadmap.

## Validation

```bash
gofmt -w ./cmd ./internal
go vet ./...
go test ./...
go build ./cmd/home-securityd
```

## Privacy and security

Camera streams, recordings, snapshots, detections, biometric features, plate data, and household activity are private by default. No remote telemetry or cloud AI path is part of the current implementation. See `PRIVACY.md` and `SECURITY.md`.

## License

No project license has been selected yet. A license decision is required before a public software release is represented as open source under a specific license.
