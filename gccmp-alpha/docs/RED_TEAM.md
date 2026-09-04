# Red-team ledger

| Claim | Current evidence | Status | Required next evidence |
|---|---|---|---|
| Same bytes produce same object identity | SHA-256 whole-file and fixed-chunk tests | locally verified | native Windows/macOS convergence |
| mtime is not causality | mtime is excluded; regression test changes mtime only | locally verified | native runner repetition |
| unique same-content path change is detectable | deterministic `RENAMED` relation | locally verified | larger adversarial corpus |
| rename history is knowable from snapshots | impossible in general | rejected | append-only operation/parent journal |
| non-ASCII names are portable | no full normalization profile exists in v0 | unverified | explicit Unicode profile and corpus |
| source trees cannot be damaged | implementation is read-only by construction | locally observed | syscall/file-access tracing on native runners |
| cloud listings define truth | provider listings may be stale or incomplete | rejected | immutable commit graph + backend conformance |
| alpha is a filesystem | no projection/mount exists | rejected | later platform adapter phase |
