# GCCmp alpha staging branch

This branch temporarily hosts the Windows/macOS Phase-0 comparator proof under `gccmp-alpha/`.

The intended module identity is `github.com/timelabs-cpo/gccmp`, but no connected GitHub account or repository named `timelabs-cpo` was available when this branch was created. This branch is therefore an **incubator and CI evidence surface**, not the final repository location.

Do not merge the unrelated historical `1/` and `1.xcodeproj/` material into GCCmp. When the target repository exists, move only:

- `gccmp-alpha/**` to repository root;
- `.github/workflows/gccmp-alpha-ci.yml` to `.github/workflows/ci.yml`;
- `.github/workflows/gccmp-alpha-nightly.yml` to `.github/workflows/nightly.yml`.

Preserve the commit and CI receipts during migration.
