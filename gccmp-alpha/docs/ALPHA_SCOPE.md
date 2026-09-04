# Alpha scope and claim ceiling

## Research question

Can independently built Windows and macOS binaries transform the same frozen directory fixtures into byte-identical observations and comparison receipts without relying on timestamps or implicit authority?

## Alpha hypothesis

For the admitted v0 domain—regular files, directories, UTF-8 relative paths, exact bytes—the same input tree produces the same canonical snapshot payload on Windows and macOS. Two verified snapshots produce the same comparison payload on every supported runner.

## Falsifiers

The alpha fails if any of the following occurs:

1. identical frozen fixtures produce different canonical bytes on different operating systems;
2. changing only file modification time changes a snapshot;
3. ambiguous same-content rename candidates are assigned a unique history;
4. a tampered or unknown-schema snapshot is accepted;
5. unsupported path/file semantics disappear from the receipt;
6. a comparison claims causal ordering from snapshot timestamps;
7. source trees are mutated by any command.

## Deliberate limitations

- no cloud APIs or drive crawling;
- no journal/parent graph, therefore no `SUPERSEDES` claim;
- no Unicode normalization or full case folding in v0; non-ASCII portability is explicit uncertainty;
- no symlink equivalence claim;
- no operating-system namespace projection;
- no automatic merge;
- no long-running daemon;
- no LLM in the authority path.

## Exit toward beta

The next phase starts only after native Windows and macOS runs produce byte-identical frozen-fixture outputs. It adds an append-only operation journal and parent IDs so `CONCURRENT` can be distinguished from merely `DIFFERENT`.
