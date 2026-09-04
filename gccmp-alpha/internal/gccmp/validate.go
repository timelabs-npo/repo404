package gccmp

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const snapshotCausalityV01 = "unavailable:snapshot-has-no-parent-operation-log"

func validateSnapshotEnvelope(env SnapshotEnvelope) error {
	if env.Schema != SnapshotEnvelopeSchema {
		return fmt.Errorf("unsupported snapshot envelope schema %q", env.Schema)
	}
	if env.Payload.Schema != SnapshotSchema {
		return fmt.Errorf("unsupported snapshot payload schema %q", env.Payload.Schema)
	}
	if !validSHA256ID(env.PayloadSHA256) {
		return fmt.Errorf("invalid snapshot payload digest %q", env.PayloadSHA256)
	}
	if err := validateSnapshotPayload(env.Payload); err != nil {
		return err
	}
	payload, err := canonicalJSON(env.Payload)
	if err != nil {
		return err
	}
	if got := digestBytes(payload); got != env.PayloadSHA256 {
		return fmt.Errorf("snapshot payload digest mismatch: got %s want %s", got, env.PayloadSHA256)
	}
	return nil
}

func validateSnapshotPayload(payload SnapshotPayload) error {
	if payload.Label == "" {
		return fmt.Errorf("snapshot label must be non-empty")
	}
	if payload.ChunkSize != DefaultChunkSize {
		return fmt.Errorf("unsupported chunk size %d", payload.ChunkSize)
	}
	if payload.Causality != snapshotCausalityV01 {
		return fmt.Errorf("unsupported causality declaration %q", payload.Causality)
	}

	previousPath := ""
	for i, entry := range payload.Entries {
		if err := validateEntry(entry, payload.ChunkSize); err != nil {
			return fmt.Errorf("entry %d (%q): %w", i, entry.Path, err)
		}
		if i > 0 && entry.Path <= previousPath {
			return fmt.Errorf("entries are not in strict path order or contain duplicate path %q", entry.Path)
		}
		previousPath = entry.Path
	}

	for i, issue := range payload.PortabilityIssues {
		if issue.Code == "" || issue.Path == "" || issue.Detail == "" {
			return fmt.Errorf("portability issue %d is incomplete", i)
		}
		if i > 0 {
			prev := payload.PortabilityIssues[i-1]
			if issue.Path < prev.Path || (issue.Path == prev.Path && issue.Code < prev.Code) {
				return fmt.Errorf("portability issues are not canonically ordered")
			}
		}
	}

	expectedIssues := make([]PortabilityIssue, 0)
	for _, entry := range payload.Entries {
		_, issues := portableName(entry.Path)
		expectedIssues = append(expectedIssues, issues...)
		switch entry.Kind {
		case "symlink":
			expectedIssues = append(expectedIssues, PortabilityIssue{
				Code:   "unsupported_symlink",
				Path:   entry.Path,
				Detail: "symlink recorded but not followed; snapshots containing it cannot claim portable namespace equivalence",
			})
		case "other":
			expectedIssues = append(expectedIssues, PortabilityIssue{
				Code: "unsupported_file_kind", Path: entry.Path, Detail: entry.UnsupportedNote,
			})
		}
	}
	sort.Slice(expectedIssues, func(i, j int) bool {
		if expectedIssues[i].Path != expectedIssues[j].Path {
			return expectedIssues[i].Path < expectedIssues[j].Path
		}
		return expectedIssues[i].Code < expectedIssues[j].Code
	})
	if !reflect.DeepEqual(payload.PortabilityIssues, expectedIssues) {
		return fmt.Errorf("portability issues do not match deterministic derivation from entries")
	}
	expectedConflicts := findPortabilityConflicts(payload.Entries)
	if !reflect.DeepEqual(payload.PortabilityConflicts, expectedConflicts) {
		return fmt.Errorf("portability conflicts do not match deterministic derivation from entries")
	}

	previousKey := ""
	for i, conflict := range payload.PortabilityConflicts {
		if conflict.PortableKey == "" || len(conflict.Paths) < 2 {
			return fmt.Errorf("portability conflict %d is incomplete", i)
		}
		if i > 0 && conflict.PortableKey <= previousKey {
			return fmt.Errorf("portability conflicts are not in strict key order")
		}
		previousKey = conflict.PortableKey
		previousConflictPath := ""
		for j, path := range conflict.Paths {
			if !validRelativeSnapshotPath(path) {
				return fmt.Errorf("portability conflict %d has invalid path %q", i, path)
			}
			if j > 0 && path <= previousConflictPath {
				return fmt.Errorf("portability conflict %d paths are not strictly ordered", i)
			}
			previousConflictPath = path
		}
	}
	return nil
}

func validateEntry(entry Entry, chunkSize int) error {
	if !validRelativeSnapshotPath(entry.Path) {
		return fmt.Errorf("invalid relative snapshot path")
	}

	switch entry.PortableStatus {
	case "verified-ascii-v0":
		if !isASCII(entry.Path) {
			return fmt.Errorf("ASCII portability status used for non-ASCII path")
		}
		key, _ := portableName(entry.Path)
		if key == "" || entry.PortableKey != key {
			return fmt.Errorf("portable key mismatch")
		}
	case "unknown-non-ascii-v0":
		if isASCII(entry.Path) || entry.PortableKey != "" {
			return fmt.Errorf("non-ASCII uncertainty status is inconsistent")
		}
	case "unsupported":
		// Unsupported filesystem kinds may still carry the portable key that
		// was computed before kind-specific semantics were rejected.
	default:
		return fmt.Errorf("unknown portable status %q", entry.PortableStatus)
	}

	switch entry.Kind {
	case "file":
		if entry.Size < 0 {
			return fmt.Errorf("negative file size")
		}
		if !validSHA256ID(entry.SHA256) {
			return fmt.Errorf("invalid file digest")
		}
		if entry.UnsupportedNote != "" {
			return fmt.Errorf("regular file carries unsupported note")
		}
		if entry.Size == 0 {
			if len(entry.Chunks) != 0 {
				return fmt.Errorf("empty file must have no chunks")
			}
			return nil
		}
		if len(entry.Chunks) == 0 {
			return fmt.Errorf("non-empty file has no chunks")
		}
		var expectedOffset int64
		for i, chunk := range entry.Chunks {
			if chunk.Offset != expectedOffset {
				return fmt.Errorf("chunk %d offset %d; expected %d", i, chunk.Offset, expectedOffset)
			}
			if chunk.Length <= 0 || chunk.Length > chunkSize {
				return fmt.Errorf("chunk %d has invalid length %d", i, chunk.Length)
			}
			if i < len(entry.Chunks)-1 && chunk.Length != chunkSize {
				return fmt.Errorf("non-final chunk %d is not exactly %d bytes", i, chunkSize)
			}
			if !validSHA256ID(chunk.SHA256) {
				return fmt.Errorf("chunk %d has invalid digest", i)
			}
			length := int64(chunk.Length)
			if expectedOffset > entry.Size-length {
				return fmt.Errorf("chunk %d exceeds declared file size", i)
			}
			expectedOffset += length
		}
		if expectedOffset != entry.Size {
			return fmt.Errorf("chunk coverage %d does not equal file size %d", expectedOffset, entry.Size)
		}
	case "directory":
		if entry.Size != 0 || entry.SHA256 != "" || len(entry.Chunks) != 0 || entry.UnsupportedNote != "" {
			return fmt.Errorf("directory carries file or unsupported metadata")
		}
	case "symlink", "other":
		if entry.PortableStatus != "unsupported" || entry.UnsupportedNote == "" {
			return fmt.Errorf("unsupported filesystem kind lacks explicit unsupported state")
		}
		if entry.Size != 0 || entry.SHA256 != "" || len(entry.Chunks) != 0 {
			return fmt.Errorf("unsupported filesystem kind carries file content metadata")
		}
	default:
		return fmt.Errorf("unknown entry kind %q", entry.Kind)
	}
	return nil
}

func validRelativeSnapshotPath(path string) bool {
	if path == "" || strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") || strings.ContainsRune(path, '\x00') {
		return false
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func validSHA256ID(value string) bool {
	const prefix = "sha256:"
	if len(value) != len(prefix)+64 || !strings.HasPrefix(value, prefix) {
		return false
	}
	for _, ch := range value[len(prefix):] {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return true
}
