# GoreeCloud Home Security — Architecture

## Design principles

- Original GoreeCloud application architecture and local-first processing.
- Explicit trust boundaries between camera protocol/media parsing, authoritative orchestration/metadata, inference, UI/API, and platform adapters.
- Bounded supporting dependencies rather than a complete third-party NVR application layer.
- Fail closed on malformed configuration/protocol data, unsafe credential transfer, unknown security/privacy evidence, and unsupported contracts.

## Current process boundary

```text
home-securityd
  |-- config + secret references + event journal + API
  |-- FFprobe (bounded unauthenticated inspection)
  |
  +-- protected descriptor FD3 ----------+
  |                                       v
  |                           home-security-media-worker
  |                           RTSP/RTSPS + auth + RTP flow
  |                                       |
  +<-- sanitized status FD4 --------------+

Future bounded outputs -> recorder / live view / activity / detector / event core
```

`home-securityd` remains authoritative for configuration, policy, event/recording metadata, APIs, and orchestration. The owned media worker is a replaceable failure/resource domain and is never authoritative for household authorization, policy, retention, or review state.

## Implemented camera/session path

The parent resolves secret references and passes protected descriptor v2 over FD3. The worker performs RTSP/RTSPS DESCRIBE/SETUP/PLAY and emits `media_ready` over FD4 only after receiving video RTP on the negotiated TCP-interleaved channel. Digest auth is supported; Basic is RTSPS-only. TLS verification is not bypassed. Strict parser/resource bounds and sanitized categorical errors limit exposure of untrusted camera input.

Supervisor state is transient and uses capped restart backoff. Authentication/policy failures stop retrying; connectivity/stall failures may retry. The current controlled integration fixture proves the protocol path against a local test server, not against physical camera hardware.

## Planned downstream architecture

The media worker will later feed separately bounded recorder, live-view/restream, and activity-analysis paths. Motion gating will reduce detector workload; detector workers remain replaceable CPU/GPU/NPU backends. Detection/tracking feeds zones/rules and the authoritative event core, which then drives review, export, and minimized alert/platform events.

Recorder media and metadata must remain separately manageable. The existing JSONL event journal is a Development foundation, not the final transactional metadata database. No recording segments/indexes are created today.

## Platform boundary

GoreeCloud Identity, Wardveil Security, Privacy Shield, Everkeep, Mesh, Manager, and Glaze UI remain required/applicable and unaccepted. The loopback-only API and Development/nonconformant lifecycle remain in force until substantive integrations and exact evidence exist.
