# GoreeCloud Home Security — Media Engine

## Status

Development / source-and-controlled-integration validated. The media plane contains bounded probing, protected credential transfer, a GoreeCloud-owned authenticated RTSP/RTSPS session worker, supervision/backoff, and plan-only recording. It does **not** establish exact real-camera interoperability, working NVR recording, or production acceptance.

## Protected worker contract

The authenticated media boundary uses descriptor schema version `2`:

- parent resolves configured camera username/password references;
- strict descriptor travels over inherited `--descriptor-fd=3`;
- worker emits only sanitized events over `--status-fd=4`;
- worker argv never contains a camera URL, username, or password;
- worker environment is fixed to `LANG=C`, `LC_ALL=C`, and `TZ=UTC`;
- descriptor contains a bounded session read/stall timeout and is strict size/field validated.

This reduces ordinary local secret exposure but is not protection against privileged host/process-memory inspection. Production packaging must additionally trust the exact worker executable and its path/integrity.

## GoreeCloud-owned RTSP worker

`home-security-media-worker` consumes the descriptor directly and implements the current session protocol without reconstructing credential-bearing media-tool arguments. It performs:

1. RTSP or RTSPS TCP connection.
2. `DESCRIBE` with supported authentication challenge handling.
3. Bounded SDP parsing and first video-track control selection.
4. `SETUP` using `RTP/AVP/TCP;unicast;interleaved=0-1` and validation of the returned interleaved video channel.
5. `PLAY`.
6. Interleaved RTP/RTCP draining and bounded media-flow stall detection.
7. `media_ready` on FD4 only after the first packet on the negotiated video RTP channel.

Supported authentication is Digest MD5, MD5-sess, SHA-256, and SHA-256-sess with `qop=auth`. Basic authentication is allowed only when transport is RTSPS. Plain RTSP + Basic fails closed as `rtsp_basic_insecure`.

RTSPS uses normal certificate/hostname verification and TLS >=1.2. There is deliberately no `InsecureSkipVerify` path. A private-CA/self-signed camera therefore needs a future explicit trust-store design.

RTSP parsing is bounded: line/header/body counts and sizes, SDP/control URLs, transport/session/auth challenge values, and descriptor/status payloads are constrained. Folded headers fail closed. Raw protocol/network errors are converted to categorical session reasons.

## Supervision semantics

`media_sessions_enabled` remains `false` by default. When enabled, each configured camera is supervised with capped exponential backoff. Credential resolution/authentication/policy failures are blocked without an automatic retry loop; connectivity/session-stall failures may retry. Session `running` now means video RTP data has arrived, not merely that a child process started.

The controlled integration test uses a local RTSP server and verifies Digest-authenticated DESCRIBE/SETUP/PLAY through first interleaved RTP readiness. This is not physical-camera or long-duration evidence.

## FFprobe and FFmpeg boundary

FFprobe remains a bounded optional probe adapter for unauthenticated definitions. FFmpeg remains a bounded foundation for the plan-only future recording command. Credentialed long-running sessions no longer require constructing a credential-bearing FFmpeg argv.

No FFmpeg/FFprobe production package version, container digest, hardware-acceleration build, or distribution profile is approved/pinned yet.

## Recording boundary

The recording-plan builder validates camera-scoped paths and segment duration, but `home-securityd` does not execute it. There is no crash-safe segment index, retention deletion, storage-pressure policy, playback API, or unattended NVR recording yet.

## Next milestones

1. Exact physical-camera sustained-flow/progress/stall/reconnect and RTSPS trust validation.
2. Persisted camera offline/recovery event transitions; tamper remains a distinct signal.
3. Recorder execution with crash-safe segment indexing.
4. Retention/storage-pressure enforcement before unattended recording.
5. ONVIF discovery and capabilities.
6. Local live view/restreaming.
7. Motion gating, local detector workers, tracking, zones/rules, and hardware acceleration.
