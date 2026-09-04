# System Comparator audit receipt — 2026-09-04

Audit mode: read-only inspection plus isolated evidence publication. No merge, release, deployment, user-file mutation, or cloud synchronization was performed.

## Evidence lines

### A. Public GCCmp alpha

- repository: `timelabs-npo/repo404`
- PR: `#1`
- branch: `incubator/gccmp-alpha-nightly`
- source commit: `2b0e663a330720245c596729062b8ba926b09771`
- native CI run: `33882726071` — success
- verified targets: Windows x64, macOS arm64, macOS x64
- verified cross-OS canonical hashes:
  - `snapshot-a.json` — `fa170867104274a534232642ff698606117f69e26928d204393352ca76d73e0f`
  - `snapshot-b.json` — `296b7de5ccd8f47a39eb02bc04f8fe3b17045f202f91565d98afc1aeba03bad9`
  - `comparison.json` — `401458a7cb648c6b2383009276f5a64f3915fd8279d45a8e52970b1d163f0f01`

This line is externally reproducible but deliberately limited: no causal journal, provider backend, replication, mount projection, point-in-time live scan, or security channel is proven.

### B. Rhea Comparator Pro build evidence

Attached evidence identifies:

- version: `0.1.0-alpha.1`
- tested commit: `f3dbf465b5e9bf8d0071a8ff9a510bc0942ec5d2`
- tested tree: `405500066bc737b1e98ba24b8aaad97d91f08c47`
- protocol: `rhea.comparator/0.1`

The record reports race-tested Go semantics, CAS/ref/vault invariants, authenticated API operations, peer pull, explicit two-node merge, Linux execution, cross-builds, and Apple source checks. The corresponding source repository/commit was not discoverable through the connected GitHub installation or Google Drive searches by project name, protocol, version, or commit prefix. Therefore this line has strong internal build evidence but a broken external provenance edge.

### C. Historical `rhead` name collision

Google Drive contains a 2026-02-26 `projects/rh.1/logs/rhead.log` showing a Python/Uvicorn daemon start, three HTTP 200 GET responses, and clean shutdown. No evidence currently links that daemon to the newer Go `rhead`; identical naming is not ancestry.

## Red-team verdict

The current best model is:

1. the public GCCmp branch is a small, independently verifiable observation/comparison reference;
2. a stronger local runtime probably existed and was tested, but its source/publish edge is missing;
3. neither line automatically supersedes the other;
4. GitHub is not yet the complete system of record;
5. compatibility must be demonstrated by an executable mapping, not inferred from naming.

The project has reproduced its own target failure mode: the strongest progress fragment is less discoverable and less independently reproducible than an older, smaller public slice.

## Highest-value next gates

1. Recover the exact advanced commit/tree, or record an explicit `provenance_break` if only file contents remain.
2. Run one frozen fixture through GCCmp and `rhea.comparator/0.1`; emit a deterministic field/relation mapping or an incompatibility receipt.
3. Execute the advanced semantic corpus natively on macOS arm64 and Windows x64.
4. Close the journal/ref/completion-receipt crash window using transactional recovery tests.
5. Define stable observation of actively changing source trees; emit `UNSTABLE_OBSERVATION` rather than canonicalizing a mixed read.
6. Add encryption, device identity/revocation, safe GC roots, and host authorization before external beta claims.

## Commercial positioning

Recommended wedge: **evidence-preserving workspace continuity for AI R&D teams**, not “a universal filesystem for every cloud.” The first sellable surface is a read-only artifact inventory and provenance/conflict graph that emits independently verifiable Git handoff receipts. Mounts, cloud authority, and autonomous mutation come later.

## Audit artifacts

Canonical local archive:

- `system-comparator-audit-2026-09-04.zip`
- SHA-256: `b0ab760b54bb9fa000dd05c5e0b2a5da3af6e2dea7838c2dbf72286e0a2c7f65`

Google Drive copies:

- Codex folder: `https://drive.google.com/drive/folders/1bKUlPt-aDcmJ-4_sbh2Y0lHHo4LRxRlb`
- Antigravity folder: `https://drive.google.com/drive/folders/1FKAhlp6uSW8iOm7T9jvUJxxi2XPBiy_5`

Each folder contains the full Russian audit, machine-readable fact ledger, live-check receipt, commercialization brief, continuation handoff, `SHA256SUMS`, and the ZIP archive.
