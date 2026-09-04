# Alpha scope and claim ceiling

## Research question

Can independently built Windows and macOS binaries transform the same frozen directory fixtures into byte-identical observations and comparison receipts without relying on timestamps or implicit authority?

## Alpha hypothesis

For the admitted v0 domain—regular files, directories, UTF-8 relative paths, exact bytes—the same frozen input tree produces the same canonical snapshot payload on Windows and macOS. Two verified snapshots produce the same comparison payload on every supported runner.

## Falsifiers

The alpha fails if any of the following occurs:

1. identical frozen fixtures produce different canonical bytes on different operating systems;
2. changing only file modification time changes a snapshot;
3. ambiguous same-content rename candidates are assigned a unique history or cause unmatched entries to disappear;
4. a tampered, noncanonical, duplicate-key, duplicate-path, trailing-data, or unknown-schema snapshot is accepted;
5. invalid chunk coverage reaches the comparator;
6. unsupported path/file semantics disappear from the receipt;
7. a conflict relation is downgraded to `changed` because of relation ordering;
8. a comparison claims causal ordering from snapshot timestamps;
9. source trees are mutated by any command;
10. build receipts conflate the tested source commit with a synthetic workflow/merge commit.

## Deliberate limitations

- no cloud APIs or drive crawling;
- no journal/parent graph, therefore no `SUPERSEDES` or `CONCURRENT` claim;
- no point-in-time source snapshot: evidence-grade input must be quiescent/frozen;
- no Unicode normalization or full case folding in v0; non-ASCII portability is explicit uncertainty;
- no symlink, hard-link, ACL, xattr, permission, sparse-file, resource-fork, or alternate-data-stream equivalence claim;
- no operating-system namespace projection;
- no automatic merge;
- no long-running daemon;
- no LLM in the authority path;
- output temp-file/rename is not yet a durable fsync-backed transaction journal.

## Exit toward the next alpha slice

After native Windows/macOS golden convergence is reproducible, add an append-only operation journal and parent IDs. Only that evidence can distinguish `CONCURRENT`, `SUPERSEDES`, and `SUBSUMED_BY` from merely `DIFFERENT`.
