# Continue from here — provenance recovery gate

## Mission

Recover the exact source/provenance of the advanced comparator implementation, then prove or reject compatibility with the public GCCmp alpha. Do not add cloud sync, mounts, UI polish, autonomous mutation, or new project names before this gate closes.

## Frozen facts

```text
Public reference:
  repo      timelabs-npo/repo404
  PR        #1
  branch    incubator/gccmp-alpha-nightly
  source    2b0e663a330720245c596729062b8ba926b09771
  CI        33882726071 PASS

Advanced build evidence:
  version   0.1.0-alpha.1
  commit    f3dbf465b5e9bf8d0071a8ff9a510bc0942ec5d2
  tree      405500066bc737b1e98ba24b8aaad97d91f08c47
  protocol  rhea.comparator/0.1
```

## Required sequence

1. Search the local Mac, mounted volumes, Codex worktrees, Antigravity workspaces, archives and `.git` object stores for the exact commit, tree, protocol and version.
2. Freeze every candidate before changing it; record path, volume, byte size, Git status, HEAD, remotes and SHA-256 manifest.
3. If exact Git objects exist, create a `git bundle` preserving the original SHA.
4. If only files remain, create a content manifest and mark `provenance_break=true`; never recreate the old SHA fictionally.
5. Run the reproduction commands recorded in the build evidence.
6. Execute the advanced semantic tests natively on macOS arm64 and Windows x64; retain receipts and binary hashes.
7. Feed one identical frozen fixture through GCCmp and the advanced runtime.
8. Produce a deterministic field/relation mapping or an explicit incompatibility report.
9. Publish to a clean standalone repository only after the source line is reconstructed.

## Acceptance

- exact advanced source is independently checkable;
- every claim maps to a command, exit status and artifact hash;
- public and advanced formats have an executable conformance result;
- timestamps/cloud listings are never causal authority;
- conflicts and uncertainty remain visible;
- recovery does not overwrite source data.

## Stop conditions

Stop and emit evidence when multiple candidate trees claim one identity, the exact commit cannot be reconstructed, native output diverges from container evidence, migration changes IDs without a format version, a worker proposes unknown binary auto-merge, or credentials/private content would enter Git.

## Commercial output

Produce the case study: **“The comparator recovered its own fragmented development lineage.”** That is a stronger product proof than another architecture diagram.
