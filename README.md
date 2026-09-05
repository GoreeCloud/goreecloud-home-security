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
- A versioned protected-worker descriptor contract that resolves credential references in the parent process and transfers username/password material through an anonymous pipe rather than command arguments or child environment variables.
- An opt-in long-running FFmpeg session supervisor for **unauthenticated** streams with direct process execution, bounded socket I/O timeout, capped exponential restart backoff, and sanitized per-camera session state.
- Media subprocesses receive a fixed minimal environment instead of inheriting the daemon environment, reducing accidental credential propagation.
- Credentialed FFprobe/FFmpeg media execution remains fail-closed until an authenticated worker actually consumes the protected descriptor contract.
- A shell-free FFmpeg recording-plan primitive with camera-scoped segment paths; the recording plan is not executed yet.
- A local append-only JSONL event journal with restrictive file permissions and fail-closed reads.
- `GET /healthz`, `GET /readyz`, `GET /api/v1/status`, `GET /api/v1/cameras`, and `GET /api/v1/events`.
- Request IDs, no-store response policy, structured API errors, bounded event pagination, and sanitized aggregate session counts.
- Unit tests, race tests, `go vet`, build validation, and repository-governance checks.

The long-running session supervisor is **disabled by default** (`media_sessions_enabled: false`). When explicitly enabled, it only establishes a Development RTSP/RTSPS session and copies the selected video stream to a null sink; it does not record media, provide playback/live view, perform detection, or prove camera compatibility on real hardware.

**Protected authenticated RTSP ingest, recording execution, retention enforcement, playback, live restreaming, motion detection, object inference/tracking, alerting, and the Glaze UI remain unimplemented.** The application is not Stable and is not production-approved.

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

If `ffprobe` is installed and a camera definition does not require credentials, Home Security probes that stream on the configured interval. Credentialed cameras still report `credentialed_probe_blocked` for external-media execution.

To exercise the new long-running **unauthenticated** Development session supervisor, explicitly set `media_sessions_enabled` to `true`. This is not recommended as unattended NVR operation yet because recorder execution, storage-pressure policy, retention, and target-camera validation are not implemented.

See `docs/MEDIA-ENGINE.md` for the protected credential-transfer contract and current worker boundary.

The development server intentionally rejects non-loopback listen addresses until an accepted authentication, authorization, and network-exposure boundary exists.

## Target capability direction

The planned product is a capability superset for GoreeCloud home-security use, including ONVIF discovery, protected authenticated RTSP/RTSPS ingest, managed live restreaming, motion pre-filtering, pluggable local detection, tracking, zones/rules, continuous and event-aware recording, review/timeline/export, local alerts, privacy-sensitive optional intelligence, and substantive GoreeCloud platform integrations.

See `SPECIFICATIONS.md`, `FEATURES.md`, `docs/ARCHITECTURE.md`, and `docs/MEDIA-ENGINE.md` for implementation boundaries and phased work.

## Validation

```bash
gofmt -w ./cmd ./internal
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/home-securityd
```

## Privacy and security

Camera streams, recordings, snapshots, detections, biometric features, plate data, and household activity are private by default. No remote telemetry or cloud AI path is part of the current implementation. See `PRIVACY.md` and `SECURITY.md`.

## License

No project license has been selected yet. A license decision is required before a public software release is represented as open source under a specific license.
