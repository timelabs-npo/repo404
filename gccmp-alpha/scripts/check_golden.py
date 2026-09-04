#!/usr/bin/env python3
from __future__ import annotations
import hashlib
import json
import pathlib
import sys

root = pathlib.Path(__file__).resolve().parents[1]
pairs = [
    (root / "out" / "snapshot-a.json", root / "testdata" / "golden" / "snapshot-a.json"),
    (root / "out" / "snapshot-b.json", root / "testdata" / "golden" / "snapshot-b.json"),
    (root / "out" / "comparison.json", root / "testdata" / "golden" / "comparison.json"),
]
failed = False
for actual, golden in pairs:
    a = actual.read_bytes()
    g = golden.read_bytes()
    if a != g:
        print(f"MISMATCH {actual.relative_to(root)} != {golden.relative_to(root)}", file=sys.stderr)
        failed = True
    else:
        print(f"MATCH {actual.name} sha256={hashlib.sha256(a).hexdigest()}")
    json.loads(a)
if failed:
    raise SystemExit(1)
