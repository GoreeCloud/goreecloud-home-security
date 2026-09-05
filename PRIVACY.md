# GoreeCloud Home Security — Privacy

## Status

Privacy posture: Development / incomplete. Privacy Shield integration is required but not yet accepted. This file documents requirements and current source behavior; it does not claim production Privacy Shield conformance.

## Private data categories

The product may process, depending on enabled features:

- Live camera video and optional audio.
- Recordings, clips, snapshots, and thumbnails.
- Camera/device configuration and network endpoints.
- Motion/object events and timestamps.
- Object labels, scores, tracks, zones, and rule outcomes.
- Household review state and alert history.
- Optional face templates/identity matches.
- Optional license-plate observations.
- Optional embeddings and semantic-search metadata.

## Current default behavior

- Processing is local.
- No remote telemetry path exists in the current source slice.
- No external AI provider exists.
- No face recognition, plate recognition, semantic indexing, recording execution, sustained live ingest, or alert delivery exists yet.
- The development API is loopback-only.
- Camera list/status responses do not expose stream URLs or credential references.
- Camera credentials must not be embedded in stream URLs.
- Enabled unauthenticated streams may be periodically probed by local FFprobe when it is installed; only codec names, dimensions, frame rate, probe time, and categorical state are retained in public camera status.
- Credentialed camera probing is currently blocked before external process execution so reusable passwords are not placed into command arguments.
- Raw FFprobe stderr and camera/network failure details are not propagated to public status.

## Purpose limitation

Camera media and derived security events may be processed only for explicitly documented home-security functions. The current probe exists only to establish bounded stream compatibility/health metadata. A future recording, indexing, biometric, plate, or AI feature requires its own purpose, access, retention, deletion, export, backup, and external-transmission review.

## Sensitive intelligence defaults

Face recognition, license-plate recognition, semantic embeddings, person re-identification, and cross-camera correlation must remain disabled by default until an approved privacy design and working controls exist. Enabling one capability must not silently enable another.

## Retention and deletion

The current source slice stores only event metadata supplied internally to the event journal. Media probe state is held in memory and is not a recording. The FFmpeg recording-plan primitive does not create media. A production design must separately define retention/deletion for:

- Recording segments.
- Clips and snapshots.
- Thumbnails/previews.
- Event metadata.
- Search indexes and embeddings.
- Face/plate derived data.
- Temporary exports.
- Caches.
- Backups and recovery copies.

No complete deletion claim may be made until all applicable layers have working deletion behavior and known backup limitations are represented accurately.

## Export

Portable export is required before production acceptance for user-selected media and appropriate Home Security-owned metadata. Exports containing private information must be access-controlled and temporary export files must have explicit cleanup behavior.

## Privacy-safe platform status

Manager, Mesh, Privacy Shield, or other status consumers should receive minimized state such as service health, capability availability, counts where justified, and conformance state. They must not receive raw live frames, recordings, camera credentials, private stream URLs, face templates, plate observations, or detailed household activity merely to render status.
