# GCCmp alpha

A read-only, deterministic **General Common Comparator** reference implementation.

This alpha observes two directory trees, emits content-addressed canonical snapshots, and compares them without using wall-clock time as causality and without selecting a silent winner.

## Today’s proof surface

- native single-file binaries for Windows x64, macOS arm64, and macOS Intel;
- deterministic canonical snapshot envelopes with SHA-256 payload identities;
- fixed 1 MiB chunk hashes plus whole-file hashes;
- add/remove/modify/unique-content-rename detection;
- ambiguous rename candidates preserve all additions/removals rather than losing cardinality;
- explicit portable-name conflicts and non-ASCII portability uncertainty;
- strict canonical verification: duplicate JSON keys, duplicate paths, trailing data, unknown fields, malformed chunk coverage, noncanonical bytes, and digest tampering are rejected;
- byte-identical frozen-fixture receipts across all native CI runners;
- active scheduled nightly builds with downloadable artifacts, provenance receipts, and SHA-256 inventories.

## Not claimed

This is **not yet a filesystem**, cloud synchronizer, provider authority, causal-history recovery system, conflict resolver for arbitrary binary edits, or background agent orchestrator. It does not copy, delete, rename, upload, hydrate, or mutate source data.

A pair of snapshots cannot prove which side is newer. Therefore comparison output states `causal_ordering: UNKNOWN` unless a later phase supplies a parent/operation journal.

A scan is not yet a point-in-time filesystem snapshot. If source bytes change while a file or directory tree is being read, alpha v0 has no cross-platform snapshot primitive or double-read stability proof. Use quiescent/frozen inputs for evidence-grade runs.

## Build

```sh
go test ./...
go build -trimpath -o out/gccmp ./cmd/gccmp
```

## Run

```sh
./out/gccmp snapshot --label mac --out mac.json /path/to/tree
./out/gccmp snapshot --label wd  --out wd.json  /path/to/other-tree
./out/gccmp compare --out comparison.json mac.json wd.json
./out/gccmp verify mac.json
```

## Alpha acceptance test

```sh
mkdir -p out
go run ./cmd/gccmp snapshot --label fixture-a --out out/snapshot-a.json testdata/fixture-a
go run ./cmd/gccmp snapshot --label fixture-b --out out/snapshot-b.json testdata/fixture-b
go run ./cmd/gccmp compare --out out/comparison.json out/snapshot-a.json out/snapshot-b.json
python3 scripts/check_golden.py
```

## Canonicality rule

Canonical payloads contain no absolute paths, hostnames, build times, modification times, runner IDs, or random values. Build metadata belongs in a separate noncanonical receipt with distinct `source_commit` and `workflow_commit` fields.

## License

Apache-2.0.
