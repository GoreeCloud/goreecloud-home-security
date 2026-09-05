# GoreeCloud Home Security — Media Engine

## Status

Development / source-validated only. This document describes the current media-plane boundary introduced in PR #1. It does not claim live NVR ingest, recording execution, authenticated-camera compatibility, or production FFmpeg acceptance.

## Supporting dependency boundary

GoreeCloud Home Security remains the application authority. FFmpeg/FFprobe may be used as bounded supporting media foundations for protocol probing, codec inspection, remuxing, segmentation, and later decode/transcode work. They do not own camera policy, event semantics, authorization, retention policy, review state, or GoreeCloud platform integration.

The current source assumes an `ffprobe` executable may be available on the host. No production FFmpeg package/version, container digest, hardware-acceleration build, or licensing/distribution profile has been approved or pinned yet.

## Implemented probe path

For each enabled camera, the Development daemon periodically asks the bounded probe adapter to inspect an RTSP/RTSPS stream. The adapter:

- invokes `ffprobe` directly with an argument vector; it never invokes a shell;
- uses a context deadline and classifies failures instead of exposing raw process errors;
- discards media-tool stderr rather than copying camera/network diagnostics into API status;
- parses only a bounded JSON stream description;
- retains only video/audio codec names, width, height, and frame rate for public runtime status;
- publishes only categorical failure reasons such as `dependency_unavailable`, `probe_timeout`, `probe_failed`, or `invalid_probe_output`;
- never exposes the configured stream URL or credential-environment names through the camera API.

This is connectivity/codec probing, not continuous ingest. A successful probe is not proof that a camera is suitable for sustained recording or detection workloads.

## Credentialed-camera boundary

Camera configuration continues to use environment-variable references for username/password material. The current FFprobe adapter **does not resolve or inject those credentials**. If a configured camera has credential references, probing fails closed with `credentialed_probe_blocked` before any external media process is launched.

This is deliberate. Putting a reusable camera password into an FFmpeg/FFprobe URL argument can make it observable in the local process command line. Authenticated RTSP support must use an approved protected credential-transfer/ingest design before this block is removed.

## Recording-plan primitive

The source now contains a recording-plan builder for future segment workers. It validates:

- enabled camera state;
- path-safe camera identifiers;
- bounded segment duration;
- credentialed-camera blocking under the same protected-ingest rule;
- argument-vector execution rather than shell interpolation;
- camera-scoped recording paths below `<data_dir>/recordings/<camera-id>/...`;
- stream-copy segmentation (`-c copy`) as the initial low-overhead plan where camera codecs/container compatibility allow it.

The plan is **not executed by `home-securityd`**. No recording segment is created by the current daemon, no retention deletion runs, and no recording/playback API exists yet.

## Next media milestones

1. Protected authenticated RTSP ingest without reusable credentials in externally visible process arguments.
2. Long-running worker supervision, reconnect/backoff, bounded queues, and camera offline/tamper events.
3. Segment writer execution plus crash-safe segment indexing.
4. Explicit retention/storage-pressure enforcement before unattended recording is enabled.
5. Low-latency local live-view/restream path.
6. Motion/activity gating and bounded analysis-frame extraction.
7. Local detector-worker contract and object tracking.
8. Hardware acceleration capability discovery and exact-target validation.

Every milestone remains subject to GoreeCloud Identity, Wardveil Security, Privacy Shield, Everkeep, Mesh, Manager, Glaze UI, recovery, and release evidence gates where applicable.
