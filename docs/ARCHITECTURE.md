# GoreeCloud Home Security — Architecture

## Design principles

- Original GoreeCloud application architecture.
- Local-first processing and storage.
- Explicit trust boundaries between camera ingest, media parsing, inference, authoritative metadata, API/UI, and platform adapters.
- Small bounded dependencies for protocols, codecs, databases, and inference runtimes.
- No complete third-party NVR as the permanent application layer.
- Fail closed on malformed configuration, corrupt authoritative metadata, unknown security/privacy evidence, unsupported contract versions, and unsafe credential-transfer paths.

## Target logical architecture

```text
Cameras / ONVIF devices
        |
        v
+----------------------------+
| Camera & Stream Plane      |
| probe / protected secrets  |
| sessions / reconnect       |
+-------------+--------------+
              |
      +-------+-------+
      |               |
      v               v
+-----------+    +----------------+
| Recorder  |    | Activity Plane |
| segments  |    | motion gating  |
+-----+-----+    +-------+--------+
      |                  |
      |                  v
      |          +------------------+
      |          | Inference Worker |
      |          | CPU/GPU/NPU/etc. |
      |          +---------+--------+
      |                    |
      +----------+---------+
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

The current Development source implements configuration, sanitized camera/session state, event journal, loopback read-only API, and bounded media-plane primitives:

- periodic FFprobe-based RTSP/RTSPS connectivity/codec inspection for unauthenticated stream definitions;
- a versioned anonymous-pipe protected-worker descriptor contract for future authenticated media workers;
- a fixed minimal media-child environment that prevents daemon environment inheritance;
- opt-in long-running FFmpeg process supervision for unauthenticated streams, with bounded I/O timeout and capped exponential restart backoff;
- sanitized transient session status and aggregate status counts;
- a validated shell-free FFmpeg segment-recording command plan that is not executed.

The session supervisor currently stream-copies video to a null sink. It is a process/session supervision foundation, not a recorder, live-view service, detector, or proof of real-camera sustained media flow.

## Current credential boundary

`home-securityd` owns camera configuration and secret references. Credential material can be resolved into a versioned worker descriptor and transferred through an anonymous pipe intended for inherited file descriptor 3. External FFprobe/FFmpeg adapters still reject credentialed cameras, because no protected authenticated backend consumes that descriptor yet.

Future authenticated workers must read the descriptor directly after process creation and use a media API/library path that does not reconstruct credential-bearing command arguments. Worker crashes/restarts must not make the worker authoritative for household policy, authorization, event metadata, or retention.

## Planned process boundaries

- `home-securityd` — authoritative configuration, event/recording metadata, APIs, orchestration, and policy.
- GoreeCloud media workers — isolated/supervised ingest, decode, remux/segment, and live-output workers using bounded media foundations.
- Detector workers — versioned local inference interface, replaceable by accelerator backend.
- Glaze UI client — authorized Home Security UX.

## Event pipeline direction

1. Camera source is discovered/configured and probed.
2. Protected credentials are transferred to an authenticated worker when needed.
3. Supervised ingest establishes and maintains media flow.
4. Recorder and activity analysis consume bounded stream outputs.
5. Motion gating, detector, tracker, zones/rules, event core, review index, and alert engine process only the minimum required data.

Steps 1 and part of 3 have Development source foundations. Step 2 has only the protected transfer contract; no authenticated worker exists. Recorder/activity/detection/review/alert stages remain planned.

## Storage direction

Recording media and metadata must remain separately manageable. The current JSONL event journal is a Development foundation, not the final metadata database. The recording plan reserves camera-scoped paths below `<data_dir>/recordings/<camera-id>/...`, but no recording files or indexes are created yet.

## Interoperability

Initial protocols are RTSP/RTSPS and ONVIF. Optional MQTT/Home Assistant adapters may be added later, but core operation must not require Home Assistant or another external control plane.
