package gccmp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestAmbiguousRenamePreservesCardinality(t *testing.T) {
	leftDir, rightDir := t.TempDir(), t.TempDir()
	mustWrite(t, filepath.Join(leftDir, "a.txt"), []byte("same"))
	mustWrite(t, filepath.Join(leftDir, "b.txt"), []byte("same"))
	mustWrite(t, filepath.Join(rightDir, "c.txt"), []byte("same"))
	mustWrite(t, filepath.Join(rightDir, "other.txt"), []byte("different"))
	left, _ := SnapshotDirectory(leftDir, "left", "")
	right, _ := SnapshotDirectory(rightDir, "right", "")
	cmp, _ := CompareSnapshots(left, right)
	if !hasRelation(cmp, "AMBIGUOUS_RENAME") {
		t.Fatal("expected ambiguous rename")
	}
	if relationCount(cmp, "REMOVED") != 2 || relationCount(cmp, "ADDED") != 2 {
		t.Fatalf("ambiguous mapping discarded cardinality: relations=%v", cmp.Payload.Relations)
	}
	if cmp.Payload.Overall != "conflicted" {
		t.Fatalf("overall=%s", cmp.Payload.Overall)
	}
}

func TestASCIIPathCollision(t *testing.T) {
	entries := []Entry{
		{Path: "Foo.txt", PortableKey: "foo.txt"},
		{Path: "foo.txt", PortableKey: "foo.txt"},
	}
	conflicts := findPortabilityConflicts(entries)
	if len(conflicts) != 1 {
		t.Fatalf("conflicts=%v", conflicts)
	}
	if conflicts[0].PortableKey != "foo.txt" || len(conflicts[0].Paths) != 2 {
		t.Fatalf("unexpected conflict=%v", conflicts[0])
	}
}

func TestWindowsReservedNameProfileIsHostIndependent(t *testing.T) {
	_, issues := portableName("CON.txt")
	if !hasIssue(issues, "windows_reserved_name") {
		t.Fatal("expected Windows reserved-name issue")
	}
	_, issues = portableName("dir/trailing.")
	if !hasIssue(issues, "windows_trailing_dot_or_space") {
		t.Fatal("expected Windows trailing-dot issue")
	}
	_, issues = portableName("bad*name.txt")
	if !hasIssue(issues, "windows_invalid_character") {
		t.Fatal("expected Windows invalid-character issue")
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

func TestDuplicateJSONKeyRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "duplicate.json")
	data := `{"schema":"gccmp.snapshot-envelope/v0.1","schema":"gccmp.snapshot-envelope/v0.1","payload_sha256":"sha256:0","payload":{"schema":"gccmp.snapshot/v0.1","label":"x","chunk_size":1048576,"entries":[],"portability_issues":[],"portability_conflicts":[],"causality":"x"}}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSnapshot(path); err == nil || !strings.Contains(err.Error(), "duplicate JSON object key") {
		t.Fatalf("duplicate key was not rejected explicitly: %v", err)
	}
}

func TestTrailingJSONRejected(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("x"))
	env, _ := SnapshotDirectory(dir, "trailing", "")
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := WriteCanonical(path, env); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`{"second":true}`); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSnapshot(path); err == nil || !strings.Contains(err.Error(), "trailing JSON value") {
		t.Fatalf("trailing JSON was not rejected explicitly: %v", err)
	}
}

func TestDuplicatePathRejectedEvenWithMatchingDigest(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("x"))
	valid, _ := SnapshotDirectory(dir, "duplicate-path", "")
	payload := valid.Payload
	payload.Entries = append(payload.Entries, payload.Entries[0])
	payloadBytes, _ := canonicalJSON(payload)
	forged := SnapshotEnvelope{
		Schema:        SnapshotEnvelopeSchema,
		PayloadSHA256: digestBytes(payloadBytes),
		Payload:       payload,
	}
	path := filepath.Join(t.TempDir(), "duplicate-path.json")
	if err := WriteCanonical(path, forged); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSnapshot(path); err == nil || !strings.Contains(err.Error(), "duplicate path") {
		t.Fatalf("duplicate path was not rejected: %v", err)
	}
	if _, err := CompareSnapshots(forged, valid); err == nil || !strings.Contains(err.Error(), "invalid left snapshot") {
		t.Fatalf("in-memory invalid snapshot reached comparator: %v", err)
	}
}

func TestNonCanonicalEnvelopeRejected(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("x"))
	env, _ := SnapshotDirectory(dir, "pretty", "")
	canonical, _ := canonicalJSON(env)
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, canonical, "", "  "); err != nil {
		t.Fatal(err)
	}
	pretty.WriteByte('\n')
	path := filepath.Join(t.TempDir(), "pretty.json")
	if err := os.WriteFile(path, pretty.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSnapshot(path); err == nil || !strings.Contains(err.Error(), "not canonical") {
		t.Fatalf("noncanonical envelope was not rejected: %v", err)
	}
}

func TestInvalidChunkCoverageRejected(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.txt"), []byte("abcdef"))
	valid, _ := SnapshotDirectory(dir, "bad-chunk", "")
	payload := valid.Payload
	payload.Entries[0].Chunks[0].Length--
	payloadBytes, _ := canonicalJSON(payload)
	forged := SnapshotEnvelope{
		Schema:        SnapshotEnvelopeSchema,
		PayloadSHA256: digestBytes(payloadBytes),
		Payload:       payload,
	}
	path := filepath.Join(t.TempDir(), "bad-chunk.json")
	if err := WriteCanonical(path, forged); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSnapshot(path); err == nil || !strings.Contains(err.Error(), "chunk coverage") {
		t.Fatalf("invalid chunk coverage was not rejected: %v", err)
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
	return relationCount(c, typ) > 0
}

func relationCount(c ComparisonEnvelope, typ string) int {
	n := 0
	for _, r := range c.Payload.Relations {
		if r.Type == typ {
			n++
		}
	}
	return n
}

func hasIssue(issues []PortabilityIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
