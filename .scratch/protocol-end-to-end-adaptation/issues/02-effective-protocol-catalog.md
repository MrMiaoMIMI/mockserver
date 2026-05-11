# Effective protocol catalog

Status: completed
Type: AFK

## Parent

.scratch/protocol-end-to-end-adaptation/PRD.md

## What to build

Make the protocol catalog a complete authoring contract. The management API should return effective operators for every field and selector so frontend authoring does not duplicate backend operator defaulting.

## Acceptance criteria

- [x] Protocol catalog fields include effective operators even when the registered field relies on type defaults.
- [x] Protocol catalog selectors include effective operators derived from selector overrides or field defaults.
- [x] HTTP protocol metadata exposes request host/path selectors and HTTP request fields.
- [x] Cache protocol metadata exposes operation/key selectors and cache request fields.
- [x] Frontend types and API helpers represent the effective protocol catalog.
- [x] Tests cover effective operators for fields and selectors.

## Blocked by

- .scratch/protocol-end-to-end-adaptation/issues/01-generic-selector-contract.md

