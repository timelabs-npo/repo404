# Nightly build contract

The scheduled workflow builds and executes the same frozen acceptance corpus on:

- Windows Server 2025 x64;
- macOS 26 arm64;
- macOS 26 Intel x64.

Each runner uploads:

- the native `gccmp` binary;
- `snapshot-a.json`;
- `snapshot-b.json`;
- `comparison.json`;
- `BUILD-RECEIPT.json`;
- `SHA256SUMS`.

A convergence job downloads all three artifact sets and fails unless every canonical JSON receipt is byte-identical across operating systems and matches the committed golden files.

Nightly artifacts are unsigned research builds. A successful workflow proves native compilation, native test execution, and frozen-corpus convergence on the named runner images. It does not prove arbitrary-drive compatibility or signing/notarization readiness.
