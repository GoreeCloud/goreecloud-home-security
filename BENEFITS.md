# GoreeCloud Home Security — Benefits

These are product objectives where implementation is still planned.

## Current foundation benefits

- Starts from an original GoreeCloud-owned architecture rather than a complete-product fork.
- Defaults the Development API to loopback instead of silently exposing an unauthenticated camera-control surface.
- Keeps camera credentials out of stream URLs and API responses.
- Establishes an intentionally small, auditable Go service with no third-party Go dependencies in the first source slice.
- Preserves event metadata locally with explicit restrictive permissions and corruption detection.
- Separates implemented behavior from the broader NVR roadmap.

## Target product benefits

- Local-first camera processing without mandatory cloud inference or hosted control plane.
- Hardware-flexible detection with bounded accelerator adapters.
- Privacy-sensitive features that are explicit and optional rather than silently enabled.
- First-party GoreeCloud Identity, Privacy Shield, Wardveil, Everkeep, Mesh, Manager, Notify, and Glaze UI integration.
- Standards-based interoperability with common cameras and home-automation systems without making those external systems architectural authorities.
- Portable recordings, event metadata, configuration, and recovery paths under user control.
