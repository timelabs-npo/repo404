package gccmp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSnapshotDeterministicAndMtimeIndependent(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")
	mustWrite(t, p, []byte("alpha\n"))
	a, err := SnapshotDirectory(dir, "fixture", "")
	if err != nil {
		t.Fatal(err)
	}
	old := time.Unix(1, 0)
	if err := os.Chtimes(p, old, old); err != nil {
		t.Fatal(err)
	}
	b, err := SnapshotDirectory(dir, "fixture", "")
	if err != nil {
		t.Fatal(err)
	}
	ab, _ := canonicalJSON(a)
	bb, _ := canonicalJSON(b)
	if string(ab) != string(bb) {
		t.Fatalf("snapshot changed after mtime-only mutation\n%s\n%s", ab, bb)
	}
}

func TestCompareAddModifyRename(t *testing.T) {
	leftDir, rightDir := t.TempDir(), t.TempDir()
	mustWrite(t, filepath.Join(leftDir, "old.txt"), []byte("same\n"))
	mustWrite(t, filepath.Join(leftDir, "mod.txt"), []byte("before\n"))
	mustWrite(t, filepath.Join(leftDir, "gone.txt"), []byte("gone\n"))
	mustWrite(t, filepath.Join(rightDir, "new-name.txt"), []byte("same\n"))
	mustWrite(t, filepath.Join(rightDir, "mod.txt"), []byte("after\n"))
	mustWrite(t, filepath.Join(rightDir, "added.txt"), []byte("added\n"))
	left, _ := SnapshotDirectory(leftDir, "left", "")
	right, _ := SnapshotDirectory(rightDir, "right", "")
	cmp, err := CompareSnapshots(left, right)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"RENAMED": false, "MODIFIED": false, "REMOVED": false, "ADDED": false}
	for _, r := range cmp.Payload.Relations {
		if _, ok := want[r.Type]; ok {
			want[r.Type] = true
		}
	}
	for typ, seen := range want {
		if !seen {
			t.Errorf("missing relation %s", typ)
		}
	}
	if cmp.Payload.CausalOrdering != "UNKNOWN" {
		t.Fatal("snapshot comparison must not invent causality")
	}
}

func TestAmbiguousRename(t *testing.T) {
	leftDir, rightDir := t.TempDir(), t.TempDir()
	mustWrite(t, filepath.Join(leftDir, "a.txt"), []byte("same"))
	mustWrite(t, filepath.Join(leftDir, "b.txt"), []byte("same"))
	mustWrite(t, filepath.Join(rightDir, "c.txt"), []byte("same"))
	left, _ := SnapshotDirectory(leftDir, "left", "")
	right, _ := SnapshotDirectory(rightDir, "right", "")
	cmp, _ := CompareSnapshots(left, right)
	if !hasRelation(cmp, "AMBIGUOUS_RENAME") {
		t.Fatal("expected ambiguous rename")
	}
	if cmp.Payload.Overall != "conflicted" {
		t.Fatalf("overall=%s", cmp.Payload.Overall)
	}
}

func TestASCIIPathCollision(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "Foo.txt"), []byte("a"))
	mustWrite(t, filepath.Join(dir, "foo.txt"), []byte("b"))
	env, err := SnapshotDirectory(dir, "case", "")
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		return
	}
	if len(env.Payload.PortabilityConflicts) != 1 {
		t.Fatalf("conflicts=%v", env.Payload.PortabilityConflicts)
	}
}

func TestWindowsReservedName(t *testing.T) {
	dir := t.TempDir()
	name := "CON.txt"
	if runtime.GOOS == "windows" {
		name = "safe.txt"
	}
	mustWrite(t, filepath.Join(dir, name), []byte("x"))
	env, err := SnapshotDirectory(dir, "reserved", "")
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && !hasIssue(env.Payload.PortabilityIssues, "windows_reserved_name") {
		t.Fatal("expected Windows reserved-name issue")
	}
}

func TestNonASCIIIsExplicitlyUnverified(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "café.txt"), []byte("x"))
	env, err := SnapshotDirectory(dir, "unicode", "")
	if err != nil {
		t.Fatal(err)
	}
	if !hasIssue(env.Payload.PortabilityIssues, "unicode_normalization_unverified") {
		t.Fatal("non-ASCII portability uncertainty was not surfaced")
	}
}

func TestTamperDetected(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("x"))
	env, _ := SnapshotDirectory(dir, "tamper", "")
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := WriteCanonical(path, env); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	data = []byte(strings.Replace(string(data), "tamper", "tamper2", 1))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSnapshot(path); err == nil {
		t.Fatal("tampered snapshot accepted")
	}
}

func TestUnknownFieldRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	data := `{"schema":"gccmp.snapshot-envelope/v0.1","payload_sha256":"sha256:0","payload":{"schema":"gccmp.snapshot/v0.1","label":"x","chunk_size":1048576,"entries":[],"portability_issues":[],"portability_conflicts":[],"causality":"x"},"surprise":true}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSnapshot(path); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestChunking(t *testing.T) {
	dir := t.TempDir()
	data := make([]byte, DefaultChunkSize+17)
	for i := range data {
		data[i] = byte(i % 251)
	}
	mustWrite(t, filepath.Join(dir, "large.bin"), data)
	env, err := SnapshotDirectory(dir, "chunk", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(env.Payload.Entries[0].Chunks); got != 2 {
		t.Fatalf("chunks=%d", got)
	}
}

func TestCanonicalRoundTrip(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("alpha"))
	env, _ := SnapshotDirectory(dir, "roundtrip", "")
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := WriteCanonical(path, env); err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(loaded)
	want, _ := json.Marshal(env)
	if string(got) != string(want) {
		t.Fatal("round trip changed value")
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func hasRelation(c ComparisonEnvelope, typ string) bool {
	for _, r := range c.Payload.Relations {
		if r.Type == typ {
			return true
		}
	}
	return false
}

func hasIssue(issues []PortabilityIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
