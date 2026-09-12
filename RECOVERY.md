# GoreeCloud Home Security — Backup, Recovery, and Portability

## Status

Recovery posture: Required / not yet accepted. Everkeep integration and restore evidence are not implemented.

## Data classes

The final recovery model must distinguish:

1. Source code and release artifacts.
2. Application configuration.
3. Camera secret references and separately protected secret material.
4. Event/detection metadata.
5. Recording segments.
6. Clips, snapshots, and thumbnails.
7. Search indexes/embeddings and other rebuildable derived data.
8. Review state, rules, retention settings, and user-facing preferences.
9. Audit/security evidence required for incident or recovery investigation.

## Current source slice

The current event journal is a local JSONL file at `<data_dir>/events.jsonl`. It is portable as a file but there is not yet an application-level backup, restore, migration, or export command. Therefore recovery is not accepted.

## Required recovery design

- Back up configuration without silently copying reusable camera secrets into ordinary backup payloads.
- Define whether bulk recordings are protected by Everkeep, replicated through another approved storage strategy, or intentionally excluded according to user policy and capacity.
- Preserve metadata/media referential integrity across backup and restore.
- Rebuild derived indexes where practical instead of treating them as irreplaceable authority.
- Validate restore on a clean isolated environment.
- Record backup-version compatibility and migration requirements.
- Provide rollback behavior for schema/config migrations.
- Verify that restored authorization and secret references do not grant unintended access.

## Production gate

A successful backup command is not enough. Production acceptance requires documented and successful restore validation for the exact candidate and its expected data classes.
