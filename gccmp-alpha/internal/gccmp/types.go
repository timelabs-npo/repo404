package gccmp

const (
	SnapshotSchema           = "gccmp.snapshot/v0.1"
	SnapshotEnvelopeSchema   = "gccmp.snapshot-envelope/v0.1"
	ComparisonSchema         = "gccmp.comparison/v0.1"
	ComparisonEnvelopeSchema = "gccmp.comparison-envelope/v0.1"
	DefaultChunkSize         = 1 << 20
)

type Chunk struct {
	Offset int64  `json:"offset"`
	Length int    `json:"length"`
	SHA256 string `json:"sha256"`
}

type Entry struct {
	Path            string  `json:"path"`
	Kind            string  `json:"kind"`
	Size            int64   `json:"size,omitempty"`
	SHA256          string  `json:"sha256,omitempty"`
	Chunks          []Chunk `json:"chunks,omitempty"`
	PortableKey     string  `json:"portable_key,omitempty"`
	PortableStatus  string  `json:"portable_status"`
	UnsupportedNote string  `json:"unsupported_note,omitempty"`
}

type PortabilityIssue struct {
	Code   string `json:"code"`
	Path   string `json:"path"`
	Detail string `json:"detail"`
}

type PortabilityConflict struct {
	PortableKey string   `json:"portable_key"`
	Paths       []string `json:"paths"`
}

type SnapshotPayload struct {
	Schema               string                `json:"schema"`
	Label                string                `json:"label"`
	ChunkSize            int                   `json:"chunk_size"`
	Entries              []Entry               `json:"entries"`
	PortabilityIssues    []PortabilityIssue    `json:"portability_issues"`
	PortabilityConflicts []PortabilityConflict `json:"portability_conflicts"`
	Causality            string                `json:"causality"`
}

type SnapshotEnvelope struct {
	Schema        string          `json:"schema"`
	PayloadSHA256 string          `json:"payload_sha256"`
	Payload       SnapshotPayload `json:"payload"`
}

type Relation struct {
	Type      string `json:"type"`
	LeftPath  string `json:"left_path,omitempty"`
	RightPath string `json:"right_path,omitempty"`
	LeftHash  string `json:"left_hash,omitempty"`
	RightHash string `json:"right_hash,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

type Count struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type ComparisonPayload struct {
	Schema             string     `json:"schema"`
	LeftPayloadSHA256  string     `json:"left_payload_sha256"`
	RightPayloadSHA256 string     `json:"right_payload_sha256"`
	Overall            string     `json:"overall"`
	Relations          []Relation `json:"relations"`
	Counts             []Count    `json:"counts"`
	Uncertainty        []string   `json:"uncertainty"`
	CausalOrdering     string     `json:"causal_ordering"`
}

type ComparisonEnvelope struct {
	Schema        string            `json:"schema"`
	PayloadSHA256 string            `json:"payload_sha256"`
	Payload       ComparisonPayload `json:"payload"`
}
