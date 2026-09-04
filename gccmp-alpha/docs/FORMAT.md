# Canonical formats v0.1

## Snapshot

A snapshot is an envelope containing:

- `schema = gccmp.snapshot-envelope/v0.1`;
- `payload_sha256`, calculated over the exact canonical JSON bytes of `payload` plus one LF;
- a payload with a stable logical label, fixed chunk size, path-sorted entries, sorted portability findings, and an explicit causality limitation.

The snapshot excludes mtime, ctime, permissions, absolute source path, host identity, and scan time. Those properties are either platform-specific or unable to prove causality.

## Comparison

A comparison commits to the two snapshot payload identities. Relations are sorted by `(type,left_path,right_path)`.

Current relation vocabulary:

- `IDENTICAL`
- `ADDED`
- `REMOVED`
- `MODIFIED`
- `KIND_CHANGED`
- `RENAMED`
- `AMBIGUOUS_RENAME`
- `PORTABILITY_CONFLICT`
- `UNSUPPORTED`

`RENAMED` means only that a unique unmatched file on each side has the same content identity. It does not prove which filesystem operation occurred.

## Portable-name profile v0

The v0 portable key is deliberately narrow: ASCII segment case-folding plus Windows reserved-name, invalid-character, and trailing-dot/space checks. Non-ASCII paths are preserved exactly but receive `unicode_normalization_unverified` rather than an invented portability verdict.
