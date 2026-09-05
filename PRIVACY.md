# GoreeCloud Home Security — Privacy

## Status

Privacy posture: Development / incomplete. Privacy Shield integration is required but not yet accepted. This file documents requirements and current source behavior; it does not claim production Privacy Shield conformance.

## Private data categories

The product may process live camera video/audio, recordings and derived media, camera/device endpoints, event metadata, object/zone information, review/alert history, and optional future biometric/plate/semantic data.

## Current default behavior

- Processing is local.
- No remote telemetry path or external AI provider exists.
- The development API is loopback-only.
- Camera list/status responses do not expose stream URLs or credential references.
- Camera credentials must not be embedded in stream URLs.
- Enabled unauthenticated streams may be periodically probed by local FFprobe when installed; only bounded codec/dimension/FPS/probe state is retained in public status.
- Long-running FFmpeg sessions are **disabled by default**. If explicitly enabled, they currently accept only unauthenticated camera definitions and copy the selected video stream to a null sink; no media is persisted by that session path.
- Session status contains only bounded operational state, attempt/timestamp data, and categorical reasons.
- FFprobe/FFmpeg subprocesses receive a fixed minimal environment rather than the daemon environment.
- Credentialed external-media execution remains blocked. The new protected-worker descriptor can carry credentials over an anonymous pipe, but no authenticated media backend consumes it yet.
- Raw media-tool stderr and camera/network failure details are not propagated to public status.
- No face recognition, plate recognition, semantic indexing, recording execution, playback, or alert delivery exists yet.

## Purpose limitation

Current media probing and opt-in null-sink session supervision exist only to establish bounded local stream compatibility/session behavior. A future recording, indexing, biometric, plate, or AI feature requires its own purpose, access, retention, deletion, export, backup, and external-transmission review.

## Sensitive intelligence defaults

Face recognition, license-plate recognition, semantic embeddings, person re-identification, and cross-camera correlation must remain disabled by default until an approved privacy design and working controls exist. Enabling one capability must not silently enable another.

## Retention and deletion

The current source stores event metadata supplied to the JSONL journal. Probe and session status are transient in-memory operational state. The long-running null-sink session does not intentionally store media, and the FFmpeg recording-plan primitive does not create media.

A production design must separately define retention/deletion for recording segments, clips/snapshots, thumbnails, event metadata, search indexes, biometric/plate data, temporary exports, caches, and backup/recovery copies. No complete deletion claim may be made until all applicable layers have working deletion behavior and known backup limitations are represented accurately.

## Export

Portable export is required before production acceptance for user-selected media and appropriate Home Security-owned metadata. Exports containing private information must be access-controlled and temporary export files must have explicit cleanup behavior.

## Privacy-safe platform status

Manager, Mesh, Privacy Shield, or other status consumers should receive minimized state such as service health, capability availability, counts where justified, and conformance state. They must not receive raw live frames, recordings, camera credentials, private stream URLs, face templates, plate observations, or detailed household activity merely to render status.
