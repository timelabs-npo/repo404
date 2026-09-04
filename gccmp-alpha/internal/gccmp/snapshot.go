package gccmp

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

var windowsReserved = map[string]struct{}{
	"CON": {}, "PRN": {}, "AUX": {}, "NUL": {},
	"COM1": {}, "COM2": {}, "COM3": {}, "COM4": {}, "COM5": {}, "COM6": {}, "COM7": {}, "COM8": {}, "COM9": {},
	"LPT1": {}, "LPT2": {}, "LPT3": {}, "LPT4": {}, "LPT5": {}, "LPT6": {}, "LPT7": {}, "LPT8": {}, "LPT9": {},
}

func SnapshotDirectory(root, label, excludePath string) (SnapshotEnvelope, error) {
	if label == "" {
		return SnapshotEnvelope{}, fmt.Errorf("label must be non-empty")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return SnapshotEnvelope{}, err
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return SnapshotEnvelope{}, err
	}
	if !info.IsDir() {
		return SnapshotEnvelope{}, fmt.Errorf("root is not a directory: %s", root)
	}

	var absExclude string
	if excludePath != "" && excludePath != "-" {
		absExclude, _ = filepath.Abs(excludePath)
	}

	payload := SnapshotPayload{
		Schema:               SnapshotSchema,
		Label:                label,
		ChunkSize:            DefaultChunkSize,
		Entries:              make([]Entry, 0),
		Causality:            "unavailable:snapshot-has-no-parent-operation-log",
		PortabilityIssues:    make([]PortabilityIssue, 0),
		PortabilityConflicts: make([]PortabilityConflict, 0),
	}

	err = filepath.Walk(absRoot, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == absRoot {
			return nil
		}
		if absExclude != "" && samePath(path, absExclude) {
			return nil
		}
		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !utf8.ValidString(rel) {
			return fmt.Errorf("unsupported non-UTF-8 path: %q", rel)
		}
		entry := Entry{Path: rel, PortableStatus: "verified-ascii-v0"}
		portableKey, issues := portableName(rel)
		entry.PortableKey = portableKey
		if portableKey == "" {
			entry.PortableStatus = "unknown-non-ascii-v0"
		}
		payload.PortabilityIssues = append(payload.PortabilityIssues, issues...)

		mode := fi.Mode()
		switch {
		case mode.IsRegular():
			entry.Kind = "file"
			entry.Size = fi.Size()
			hash, chunks, err := hashFile(path, DefaultChunkSize)
			if err != nil {
				return err
			}
			entry.SHA256 = hash
			entry.Chunks = chunks
		case mode.IsDir():
			entry.Kind = "directory"
		case mode&os.ModeSymlink != 0:
			entry.Kind = "symlink"
			entry.PortableStatus = "unsupported"
			entry.UnsupportedNote = "symlink semantics are not admitted in alpha v0"
			payload.PortabilityIssues = append(payload.PortabilityIssues, PortabilityIssue{
				Code:   "unsupported_symlink",
				Path:   rel,
				Detail: "symlink recorded but not followed; snapshots containing it cannot claim portable namespace equivalence",
			})
		default:
			entry.Kind = "other"
			entry.PortableStatus = "unsupported"
			entry.UnsupportedNote = mode.String()
			payload.PortabilityIssues = append(payload.PortabilityIssues, PortabilityIssue{
				Code: "unsupported_file_kind", Path: rel, Detail: mode.String(),
			})
		}
		payload.Entries = append(payload.Entries, entry)
		return nil
	})
	if err != nil {
		return SnapshotEnvelope{}, err
	}

	sort.Slice(payload.Entries, func(i, j int) bool { return payload.Entries[i].Path < payload.Entries[j].Path })
	sort.Slice(payload.PortabilityIssues, func(i, j int) bool {
		if payload.PortabilityIssues[i].Path != payload.PortabilityIssues[j].Path {
			return payload.PortabilityIssues[i].Path < payload.PortabilityIssues[j].Path
		}
		return payload.PortabilityIssues[i].Code < payload.PortabilityIssues[j].Code
	})
	payload.PortabilityConflicts = findPortabilityConflicts(payload.Entries)
	return NewSnapshotEnvelope(payload)
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	return errA == nil && errB == nil && filepath.Clean(aa) == filepath.Clean(bb)
}

func hashFile(path string, chunkSize int) (string, []Chunk, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	whole := sha256.New()
	buf := make([]byte, chunkSize)
	var chunks []Chunk
	var offset int64
	for {
		n, readErr := f.Read(buf)
		if n > 0 {
			part := buf[:n]
			_, _ = whole.Write(part)
			ch := sha256.Sum256(part)
			chunks = append(chunks, Chunk{Offset: offset, Length: n, SHA256: "sha256:" + hex.EncodeToString(ch[:])})
			offset += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", nil, readErr
		}
	}
	return "sha256:" + hex.EncodeToString(whole.Sum(nil)), chunks, nil
}

func portableName(path string) (string, []PortabilityIssue) {
	parts := strings.Split(path, "/")
	keys := make([]string, 0, len(parts))
	var issues []PortabilityIssue
	for _, part := range parts {
		if !isASCII(part) {
			issues = append(issues, PortabilityIssue{
				Code:   "unicode_normalization_unverified",
				Path:   path,
				Detail: "alpha v0 preserves UTF-8 bytes but does not claim Unicode normalization/case-fold equivalence",
			})
			return "", issues
		}
		upper := strings.ToUpper(part)
		base := upper
		if idx := strings.IndexByte(base, '.'); idx >= 0 {
			base = base[:idx]
		}
		if _, ok := windowsReserved[base]; ok {
			issues = append(issues, PortabilityIssue{Code: "windows_reserved_name", Path: path, Detail: part})
		}
		if strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			issues = append(issues, PortabilityIssue{Code: "windows_trailing_dot_or_space", Path: path, Detail: part})
		}
		for _, r := range part {
			if r < 32 || strings.ContainsRune(`<>:"\\|?*`, r) {
				issues = append(issues, PortabilityIssue{Code: "windows_invalid_character", Path: path, Detail: fmt.Sprintf("%q", r)})
				break
			}
		}
		keys = append(keys, strings.ToLower(part))
	}
	return strings.Join(keys, "/"), issues
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

func findPortabilityConflicts(entries []Entry) []PortabilityConflict {
	byKey := make(map[string][]string)
	for _, e := range entries {
		if e.PortableKey != "" {
			byKey[e.PortableKey] = append(byKey[e.PortableKey], e.Path)
		}
	}
	var out []PortabilityConflict
	for key, paths := range byKey {
		if len(paths) > 1 {
			sort.Strings(paths)
			out = append(out, PortabilityConflict{PortableKey: key, Paths: paths})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PortableKey < out[j].PortableKey })
	return out
}
