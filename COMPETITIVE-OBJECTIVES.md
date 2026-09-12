# GoreeCloud Home Security — Competitive Objectives

## Benchmark boundary

Frigate is a capability benchmark, not an upstream source, architecture template, UI template, or identity source for this project.

GoreeCloud Home Security should independently match or exceed the useful product outcomes users expect from a modern local NVR while remaining a native GoreeCloud product.

## Objectives

1. **Local AI first** — object detection and future semantic/biometric processing should work locally with no mandatory cloud inference.
2. **Efficient inference** — use motion/activity gating and workload scheduling so expensive inference is not performed on every frame unnecessarily.
3. **Strong review workflow** — make security events, detections, timelines, clips, and recordings fast to inspect and triage.
4. **Flexible recording** — support continuous, motion, event, and object-aware retention policies with explicit storage-pressure behavior.
5. **Zones and rules** — support spatial rules, masks, dwell, line crossing, label filters, and per-zone alert policy.
6. **Hardware portability** — support CPU, GPU, NPU, and accelerator backends through a GoreeCloud-owned detector contract rather than binding the product to one vendor.
7. **Privacy by default** — make local processing, minimal telemetry, explicit retention, and opt-in sensitive intelligence core behavior.
8. **Security boundary** — do not depend on network location alone as authorization; integrate GoreeCloud Identity and Wardveil before normal remote exposure.
9. **Recovery and portability** — make event data, configuration, recordings, exports, and recovery behavior understandable and independently recoverable.
10. **GoreeCloud-native integration** — integrate Mesh, Manager, Notify, Privacy Shield, Wardveil, Everkeep, Identity, and Glaze UI without making another product's internal contracts the application architecture.
11. **Interoperability without dependence** — provide ONVIF/RTSP and optional MQTT/Home Assistant compatibility through bounded adapters.
12. **Capability superset direction** — extend beyond basic NVR behavior into tamper/offline detection, storage-health awareness, household authorization, recovery evidence, and GoreeCloud-wide security coordination where justified.

## Non-objectives

- Reproducing Frigate source code.
- Copying Frigate's UI, configuration format, internal process topology, database schema, or API contracts.
- Requiring Home Assistant to operate core NVR functionality.
- Requiring public cloud inference or a hosted GoreeCloud control plane.
- Enabling face recognition, plate recognition, or semantic indexing by default.
