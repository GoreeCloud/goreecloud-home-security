# GoreeCloud Home Security — Architecture

## Design principles

- Original GoreeCloud application architecture.
- Local-first processing and storage.
- Explicit trust boundaries between camera ingest, media parsing, inference, authoritative metadata, API/UI, and platform adapters.
- Small bounded dependencies for protocols, codecs, databases, and inference runtimes.
- No complete third-party NVR as the permanent application layer.
- Fail closed on malformed configuration, corrupt authoritative metadata, unknown security/privacy evidence, and unsupported contract versions.

## Target logical architecture

```text
Cameras / ONVIF devices
        |
        v
+-----------------------+
| Camera & Stream Plane |
| discovery / probe     |
| ingest / reconnect    |
+----------+------------+
           |
     +-----+------+
     |            |
     v            v
+----------+  +----------------+
| Recorder |  | Activity Plane |
| segments |  | motion gating  |
+----+-----+  +--------+-------+
     |                 |
     |                 v
     |        +------------------+
     |        | Inference Worker |
     |        | CPU/GPU/NPU/etc. |
     |        +---------+--------+
     |                  |
     +----------+-------+
                v
       +-------------------+
       | Event / Track Core|
       | zones / rules     |
       +---------+---------+
                 |
       +---------+----------+
       |                    |
       v                    v
+--------------+     +---------------+
| Metadata DB  |     | Review/Search |
| retention    |     | index/export  |
+------+-------+     +-------+-------+
       |                     |
       +----------+----------+
                  v
        +--------------------+
        | API + Glaze UI     |
        | Identity/AuthZ     |
        +----------+---------+
                   |
        +----------+-------------------------------+
        | Platform adapters                        |
        | Privacy Shield / Wardveil / Everkeep    |
        | Mesh / Manager / Notify / Identity      |
        +------------------------------------------+
```

## Current implemented subset

Only the configuration, sanitized camera registry, event journal, and local read-only API portions exist today. No boxes above that imply media ingest, recording, detection, tracking, database indexing, UI, or platform integration should be read as implemented.

## Planned process boundaries

The design should evolve toward separable failure/resource domains:

- `home-securityd` — authoritative configuration, event/recording metadata, APIs, orchestration, policy.
- Media workers — bounded camera ingest/decode/segment processes using mature media foundations.
- Detector workers — versioned local inference interface, replaceable by accelerator backend.
- UI — GoreeCloud-owned Glaze UI client consuming authorized APIs.

A worker may crash or be restarted without becoming the authoritative source for household policy, event metadata, or authorization.

## Event pipeline direction

1. Camera source is connected and probed.
2. Decode produces analysis frames at a configured rate.
3. Motion/activity gate identifies candidate regions/time windows.
4. Detector worker performs local inference when required.
5. Tracker correlates detections over time.
6. Zones/rules determine event semantics.
7. Event core persists authoritative metadata.
8. Recorder protects the required pre/post/event media window.
9. Review index exposes the event to authorized users.
10. Alert engine publishes a minimized authorized event to local integrations.

## Storage direction

Recording media and metadata should not be one opaque store. The target storage model should support independent retention and recovery decisions for segment media, event metadata, thumbnails, exports, and rebuildable indexes.

A transactional database is expected for mature metadata. Selection remains open until ingest/event query requirements are measured; the current JSONL journal is a Development foundation, not the final database commitment.

## Interoperability

Initial protocols: RTSP/RTSPS and ONVIF. Optional MQTT/Home Assistant adapters may be added later, but core operation must not require Home Assistant or another external control plane.
