# GCCmp alpha

A read-only, deterministic **General Common Comparator** reference implementation.

This alpha observes two directory trees, emits content-addressed canonical snapshots, and compares them without using wall-clock time as causality and without selecting a silent winner.

## Today’s proof surface

- native single-file binaries for Windows x64, macOS arm64, and macOS Intel;
- deterministic canonical snapshot envelopes with SHA-256 payload identities;
- fixed 1 MiB chunk hashes plus whole-file hashes;
- add/remove/modify/rename detection;
- explicit ambiguous-rename and portable-name conflicts;
- strict manifest verification and unknown-field rejection;
- byte-identical frozen-fixture receipts across all native CI runners;
- scheduled nightly builds with downloadable artifacts and SHA-256 receipts.

## Not claimed

This is **not yet a filesystem**, cloud synchronizer, provider authority, causal-history recovery system, conflict resolver for arbitrary binary edits, or background agent orchestrator. It does not copy, delete, rename, upload, hydrate, or mutate source data.

A pair of snapshots cannot prove which side is newer. Therefore comparison output states `causal_ordering: UNKNOWN` unless a later phase supplies a parent/operation journal.

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

Canonical payloads contain no absolute paths, hostnames, build times, modification times, runner IDs, or random values. Build metadata belongs in a separate noncanonical receipt.

## License

Apache-2.0.
