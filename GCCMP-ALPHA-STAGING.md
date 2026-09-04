# GCCmp alpha staging branch

This branch temporarily hosts the Windows/macOS comparator proof under `gccmp-alpha/`.

The intended module identity is `github.com/timelabs-cpo/gccmp`, but no connected GitHub account or repository named `timelabs-cpo` was available when this branch was created. This branch is therefore an **incubator and CI evidence surface**, not the final repository location.

Do not merge the unrelated historical `1/` and `1.xcodeproj/` material into GCCmp. When the target repository exists, move only:

- `gccmp-alpha/**` to repository root;
- `.github/workflows/gccmp-alpha-ci.yml` to `.github/workflows/ci.yml`;
- `.github/workflows/gccmp-alpha-nightly.yml` to `.github/workflows/nightly.yml`.

The current incubator uses `.github/workflows/gccmp-alpha-nightly-dispatch.yml` on `repo404/main` solely because GitHub schedules workflows from the default branch. Delete that temporary dispatcher after migration; the standalone repository's own `nightly.yml` becomes authoritative.

Preserve source commits, workflow commits, native artifact hashes, and convergence receipts during migration. Do not present a synthetic PR merge commit as the tested source revision.
