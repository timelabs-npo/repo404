# Nightly build contract

The active scheduled workflow builds and executes the same frozen acceptance corpus at 02:17 UTC on:

- Windows Server 2025 x64;
- macOS 26 arm64;
- macOS 26 Intel x64.

Each runner uploads:

- the native `gccmp` binary;
- `snapshot-a.json`;
- `snapshot-b.json`;
- `comparison.json`;
- `NIGHTLY-RECEIPT.json`;
- `SHA256SUMS` covering the binary, canonical evidence, and nightly receipt.

Receipts separate:

- `source_commit`: the exact comparator branch revision checked out and compiled;
- `workflow_commit`: the default-branch scheduler revision that initiated the run.

A convergence job downloads all three artifact sets and fails unless every canonical JSON receipt is byte-identical across operating systems and matches the committed golden files.

Nightly artifacts are unsigned research builds. A successful workflow proves native compilation, native test execution, and frozen-corpus convergence on the named runner images. It does not prove arbitrary-drive compatibility, point-in-time scans of live-changing directories, cloud consistency, or signing/notarization readiness.

## Incubator activation

GitHub schedules workflows only from a repository's default branch. While the alpha source remains on the incubator branch, a temporary default-branch dispatcher checks out `incubator/gccmp-alpha-nightly`, records both commits, and runs the matrix. Delete that dispatcher when the project moves to its intended standalone repository and place the nightly workflow on that repository's default branch.
