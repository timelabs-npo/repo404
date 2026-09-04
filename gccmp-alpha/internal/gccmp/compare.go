package gccmp

import (
	"fmt"
	"sort"
)

func CompareSnapshots(left, right SnapshotEnvelope) (ComparisonEnvelope, error) {
	if left.Payload.Schema != SnapshotSchema || right.Payload.Schema != SnapshotSchema {
		return ComparisonEnvelope{}, fmt.Errorf("unsupported snapshot schema")
	}

	leftByPath := make(map[string]Entry, len(left.Payload.Entries))
	rightByPath := make(map[string]Entry, len(right.Payload.Entries))
	for _, e := range left.Payload.Entries {
		leftByPath[e.Path] = e
	}
	for _, e := range right.Payload.Entries {
		rightByPath[e.Path] = e
	}

	relations := make([]Relation, 0)
	leftOnly := make(map[string]Entry)
	rightOnly := make(map[string]Entry)

	for path, le := range leftByPath {
		re, ok := rightByPath[path]
		if !ok {
			leftOnly[path] = le
			continue
		}
		relations = append(relations, samePathRelation(le, re))
	}
	for path, re := range rightByPath {
		if _, ok := leftByPath[path]; !ok {
			rightOnly[path] = re
		}
	}

	leftByContent := indexUniqueContent(leftOnly)
	rightByContent := indexUniqueContent(rightOnly)
	consumedLeft := map[string]bool{}
	consumedRight := map[string]bool{}
	for key, leftPaths := range leftByContent {
		rightPaths := rightByContent[key]
		if len(leftPaths) == 1 && len(rightPaths) == 1 {
			lp, rp := leftPaths[0], rightPaths[0]
			le, re := leftOnly[lp], rightOnly[rp]
			relations = append(relations, Relation{
				Type: "RENAMED", LeftPath: lp, RightPath: rp,
				LeftHash: le.SHA256, RightHash: re.SHA256,
				Detail: "same file content; path changed; causality remains unknown without operation history",
			})
			consumedLeft[lp], consumedRight[rp] = true, true
		} else if len(leftPaths) > 0 && len(rightPaths) > 0 {
			// Preserve every unmatched entry below as ADDED/REMOVED. An ambiguous
			// content match is evidence of a possible rename relationship, not
			// authority to discard cardinality or choose a pairing.
			relations = append(relations, Relation{
				Type: "AMBIGUOUS_RENAME", LeftPath: joinPaths(leftPaths), RightPath: joinPaths(rightPaths),
				Detail: "multiple unmatched files share the same content hash; no unique rename mapping exists; additions and removals remain explicit",
			})
		}
	}

	for path, e := range leftOnly {
		if !consumedLeft[path] {
			relations = append(relations, Relation{Type: "REMOVED", LeftPath: path, LeftHash: e.SHA256})
		}
	}
	for path, e := range rightOnly {
		if !consumedRight[path] {
			relations = append(relations, Relation{Type: "ADDED", RightPath: path, RightHash: e.SHA256})
		}
	}

	for _, c := range left.Payload.PortabilityConflicts {
		relations = append(relations, Relation{Type: "PORTABILITY_CONFLICT", LeftPath: joinPaths(c.Paths), Detail: "left snapshot portable-key collision: " + c.PortableKey})
	}
	for _, c := range right.Payload.PortabilityConflicts {
		relations = append(relations, Relation{Type: "PORTABILITY_CONFLICT", RightPath: joinPaths(c.Paths), Detail: "right snapshot portable-key collision: " + c.PortableKey})
	}

	sort.Slice(relations, func(i, j int) bool {
		a, b := relations[i], relations[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.LeftPath != b.LeftPath {
			return a.LeftPath < b.LeftPath
		}
		return a.RightPath < b.RightPath
	})

	countsMap := map[string]int{}
	hasChange := false
	hasConflict := false
	for _, r := range relations {
		countsMap[r.Type]++
		if r.Type != "IDENTICAL" {
			hasChange = true
		}
		switch r.Type {
		case "PORTABILITY_CONFLICT", "AMBIGUOUS_RENAME", "UNSUPPORTED":
			hasConflict = true
		}
	}
	overall := "identical"
	if hasConflict {
		overall = "conflicted"
	} else if hasChange {
		overall = "changed"
	}

	counts := make([]Count, 0, len(countsMap))
	for typ, n := range countsMap {
		counts = append(counts, Count{Type: typ, Count: n})
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i].Type < counts[j].Type })

	uncertainty := []string{"causal ordering cannot be inferred from two snapshots without parent or operation history"}
	if hasCode(left.Payload.PortabilityIssues, "unicode_normalization_unverified") || hasCode(right.Payload.PortabilityIssues, "unicode_normalization_unverified") {
		uncertainty = append(uncertainty, "Unicode normalization and full case-fold equivalence are unverified in alpha v0")
	}
	if hasUnsupported(left.Payload.Entries) || hasUnsupported(right.Payload.Entries) {
		uncertainty = append(uncertainty, "one or more file kinds are recorded but unsupported for portable equivalence")
	}

	payload := ComparisonPayload{
		Schema:             ComparisonSchema,
		LeftPayloadSHA256:  left.PayloadSHA256,
		RightPayloadSHA256: right.PayloadSHA256,
		Overall:            overall,
		Relations:          relations,
		Counts:             counts,
		Uncertainty:        uncertainty,
		CausalOrdering:     "UNKNOWN",
	}
	return NewComparisonEnvelope(payload)
}

func samePathRelation(left, right Entry) Relation {
	if left.Kind != right.Kind {
		return Relation{Type: "KIND_CHANGED", LeftPath: left.Path, RightPath: right.Path, LeftHash: left.SHA256, RightHash: right.SHA256}
	}
	if left.Kind == "file" {
		if left.SHA256 == right.SHA256 && left.Size == right.Size {
			return Relation{Type: "IDENTICAL", LeftPath: left.Path, RightPath: right.Path, LeftHash: left.SHA256, RightHash: right.SHA256}
		}
		return Relation{Type: "MODIFIED", LeftPath: left.Path, RightPath: right.Path, LeftHash: left.SHA256, RightHash: right.SHA256}
	}
	if left.Kind == "directory" {
		return Relation{Type: "IDENTICAL", LeftPath: left.Path, RightPath: right.Path}
	}
	if left.UnsupportedNote == right.UnsupportedNote {
		return Relation{Type: "UNSUPPORTED", LeftPath: left.Path, RightPath: right.Path, Detail: left.UnsupportedNote}
	}
	return Relation{Type: "MODIFIED", LeftPath: left.Path, RightPath: right.Path, Detail: "unsupported file-kind metadata differs"}
}

func indexUniqueContent(entries map[string]Entry) map[string][]string {
	out := map[string][]string{}
	for path, e := range entries {
		if e.Kind == "file" && e.SHA256 != "" {
			key := e.Kind + ":" + e.SHA256
			out[key] = append(out[key], path)
		}
	}
	for key := range out {
		sort.Strings(out[key])
	}
	return out
}

func hasCode(issues []PortabilityIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func hasUnsupported(entries []Entry) bool {
	for _, e := range entries {
		if e.PortableStatus == "unsupported" {
			return true
		}
	}
	return false
}

func joinPaths(paths []string) string {
	out := ""
	for i, p := range paths {
		if i > 0 {
			out += "|"
		}
		out += p
	}
	return out
}
