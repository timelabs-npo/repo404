# Research basis for the first alpha

The landscape review found no single project that already supplies the full target. The usable design is compositional: immutable content identity, a causal commit model, a mutation journal, deterministic reconciliation, provider capability declarations, and thin operating-system projections.

This alpha deliberately implements only the earliest falsifiable layer:

1. deterministic file/chunk identities;
2. canonical directory observations;
3. explicit relation output;
4. portable-name uncertainty;
5. cross-OS golden convergence;
6. zero source mutation.

It does not implement a mount, cloud provider, replication, journal, or authority ref.

## Why Go in alpha v0

The final semantic-core language remains undecided. The research recommendation favors Rust behind a stable C ABI. For the same-day Windows/macOS proof, the Go standard library provides dependency-free native builds, deterministic JSON, filesystem traversal, and SHA-256 on all target runners. The v0 formats and tests are language-neutral. A later Rust implementation must reproduce the committed golden bytes before it can replace this reference implementation.

This is therefore a semantic validation vehicle, not a permanent language lock.
