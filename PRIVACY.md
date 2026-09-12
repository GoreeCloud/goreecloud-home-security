# GoreeCloud Home Security — Privacy

## Status

Privacy posture: Development / incomplete. Privacy Shield integration is required but not yet accepted.

## Current default behavior

- Processing is local; no remote telemetry path or external AI provider exists.
- The Development API is loopback-only and minimizes camera/media/session status.
- Camera credentials are not embedded in stream URLs, API responses, worker command lines, or worker environment variables.
- Protected worker descriptor v2 transfers credentials over inherited FD3; FD4 carries only sanitized worker state. This limits ordinary exposure but does not defeat privileged host/process-memory inspection.
- Long-running media sessions remain disabled by default. When explicitly enabled, the owned worker currently validates session media flow and intentionally does not record media.
- Digest-authenticated RTSP is supported by the owned worker; Basic credentials are refused over plaintext RTSP.
- Raw RTSP headers, SDP, URLs, authentication material, network errors, and worker stderr are not promoted into public status.
- Probe/session state is transient in memory. The durable event journal is metadata-only and now has a bounded local retention policy: `event_retention_days` defaults to 30 days and is constrained to 1–365 days. Startup performs retention before the API is exposed, and the running daemon repeats the retention sweep every six hours. A journal decode/validation or retention rewrite failure fails closed rather than silently skipping malformed history.
- Retention rewrite uses a same-directory restricted temporary file, fsync, atomic replacement, and directory fsync. Event retention time uses `ended_at` when an event has one, otherwise `started_at`.
- The recording-plan primitive does not create media.
- Face recognition, plate recognition, semantic indexing, re-identification, cross-camera correlation, recording execution, playback, and alerts remain unimplemented.

## Purpose limitation

Current probe and session-worker behavior exists to establish bounded local camera connectivity, authentication, and media-flow state. It does not authorize recording, biometric processing, cloud processing, or broader household surveillance purposes. Each later sensitive feature requires explicit purpose, access, retention, deletion, export, backup, and transmission review.

## Retention and deletion

The current Development implementation now enforces bounded deletion of expired event-journal metadata. This is a local application policy only and is not Privacy Shield acceptance, a user-facing deletion workflow, backup deletion, or evidence that every future data layer has complete lifecycle handling.

A production design must separately implement and document retention/deletion for recording segments, clips/snapshots, thumbnails, indexes/embeddings, biometric/plate data, temporary exports, caches, and recovery copies. No complete deletion claim is allowed until every applicable layer has working behavior and known backup limitations.

## Privacy-safe platform status

Manager, Mesh, Privacy Shield, and other platform consumers should receive only minimized operational state needed for their role. Raw live frames, recordings, camera credentials, private stream URLs, biometric templates, plate observations, or detailed household activity must not be propagated merely for status presentation.
