# Red-team ledger

| Claim | Current evidence surface | Status ceiling | Required next evidence |
|---|---|---|---|
| Same frozen bytes produce the same object identity | SHA-256 whole-file/fixed-chunk tests plus native cross-OS golden convergence | verified for committed fixtures | larger generated corpus and independent implementation |
| mtime is not causality | mtime is excluded; regression test changes mtime only | verified for v0 snapshot format | operation-parent journal for positive causal claims |
| unique same-content path change is detectable | deterministic `RENAMED` relation | observed relation only | operation log before calling it a historical rename |
| ambiguous content mapping preserves cardinality | ambiguity plus every unmatched add/remove remains visible | unit-tested | property-generated multiplicity corpus |
| rename history is knowable from snapshots | impossible in general | rejected | append-only operation/parent journal |
| canonical JSON is unambiguous | duplicate keys, duplicate paths, trailing values, unknown fields, noncanonical bytes rejected | unit-tested | grammar fuzzing and independent decoder |
| chunk metadata is structurally valid | offsets/lengths/digests/canonical chunk boundaries validated | unit-tested | content re-read verification and parser fuzzing |
| non-ASCII names are portable | no full normalization profile exists in v0 | unverified | explicit Unicode profile and corpus |
| a live directory scan is point-in-time consistent | no cross-platform snapshot or double-read guard yet | unverified | concurrent-writer fault test and stable-read protocol |
| source trees cannot be damaged | commands expose read-only source operations | locally observed/native exercised | syscall/file-access tracing on native runners |
| output receipt is crash-durable | temp file + rename only; no fsync journal | unverified | crash injection around write/sync/rename barriers |
| cloud listings define truth | provider listings may be stale or incomplete | rejected | immutable commit graph + backend conformance |
| alpha is a filesystem | no projection/mount exists | rejected | later platform adapter phase |
