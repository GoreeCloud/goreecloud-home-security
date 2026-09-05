# GoreeCloud Home Security — Media Engine

## Status

Development / source-validated only. The current media plane now contains bounded probe, protected credential-transfer, and opt-in supervised session primitives. It does **not** claim authenticated-camera ingest, recorder execution, real-camera interoperability, or production FFmpeg acceptance.

## Supporting dependency boundary

GoreeCloud Home Security remains the application authority. FFmpeg/FFprobe are bounded supporting media foundations for protocol probing and Development session supervision, with future roles in remuxing, segmentation, decode, and transcode work. They do not own camera policy, authorization, event semantics, retention, review state, or GoreeCloud platform integration.

No production FFmpeg package/version, container digest, hardware-acceleration build, or distribution profile has been approved or pinned yet.

## Implemented probe path

For each enabled camera, the daemon periodically asks the bounded FFprobe adapter to inspect an RTSP/RTSPS stream. The adapter invokes `ffprobe` without a shell, uses a deadline, discards stderr, strict-parses bounded JSON, and retains only sanitized codec/dimension/frame-rate/status data. Credentialed camera definitions still fail closed with `credentialed_probe_blocked` before FFprobe starts.

FFprobe also receives a fixed minimal environment rather than the daemon environment, so configured camera credential variables are not inherited by the child process.

## Protected credential-transfer contract

The source now implements a protected-worker descriptor contract for the future authenticated media process boundary:

- schema version `1`;
- camera ID and credential-free RTSP/RTSPS URL validation;
- username/password resolution from configured environment-variable references in the parent process;
- fail-closed `credential_unavailable` behavior when either secret cannot be resolved;
- strict descriptor decoding with a 64 KiB maximum payload;
- an anonymous pipe transport primitive intended to become inherited descriptor file descriptor `3`;
- fixed worker arguments containing only `--descriptor-fd=3`;
- a fixed minimal child environment (`LANG=C`, `LC_ALL=C`, `TZ=UTC`) rather than inherited application secrets.

This keeps reusable credentials off the intended worker command line and environment. It does **not** make credentials inaccessible to privileged host inspection, and it is not yet connected to an authenticated RTSP backend. The current FFprobe/FFmpeg adapters therefore continue to reject credentialed cameras.

A future authenticated worker must consume this descriptor directly through a media API/library path and must not reconstruct a credential-bearing FFmpeg command line.

## Opt-in long-running session supervision

`media_sessions_enabled` defaults to `false`. When explicitly enabled, each enabled **unauthenticated** camera can run a supervised FFmpeg session that:

- invokes FFmpeg directly without a shell;
- receives the same fixed minimal child environment;
- selects the first video stream;
- uses stream copy rather than decode;
- sends output to a null sink, so this path intentionally does not persist media;
- applies a bounded `rw_timeout` value;
- reports only `idle`, `starting`, `running`, `backoff`, `blocked`, `stopped`, or `disabled` state plus bounded attempt/timestamp/reason metadata;
- uses capped exponential restart backoff when the process exits/fails;
- discards FFmpeg stdout/stderr rather than promoting private camera/network diagnostics into API state.

`running` means the local FFmpeg process successfully started. It is not proof that media is continuously flowing, compatible with recording/detection, or validated against target hardware. Exact real-camera stall/progress monitoring remains future work.

A non-credential stream URL is still present in the Development FFmpeg process argument vector. It is private configuration and this path assumes a trusted local host administrator. Credential-bearing URLs remain prohibited.

## Recording-plan primitive

The recording-plan builder still validates enabled state, path-safe camera IDs, bounded segment duration, credentialed-camera blocking, argument-vector execution, camera-scoped paths under `<data_dir>/recordings/<camera-id>/...`, and stream-copy segmentation.

The plan is **not executed by `home-securityd`**. No segment index, retention deletion, storage-pressure handling, recording API, or playback exists yet.

## Next media milestones

1. Implement a GoreeCloud-owned authenticated media worker that consumes the protected descriptor directly without exposing credentials in argv/environment.
2. Add real-camera sustained-flow/progress monitoring and reconnect/offline/recovery event semantics; keep tamper detection separate from ordinary connectivity failure.
3. Execute segment writers with crash-safe segment indexing.
4. Enforce explicit retention and storage-pressure policy before unattended recording is enabled.
5. Add ONVIF discovery and capability inspection.
6. Add low-latency local live view/restreaming.
7. Add motion/activity gating, local detector workers, object tracking, zones/rules, and hardware-acceleration discovery.

Every milestone remains subject to GoreeCloud Identity, Wardveil Security, Privacy Shield, Everkeep, Mesh, Manager, Glaze UI, recovery, and release evidence gates where applicable.
