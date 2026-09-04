package gccmp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

func canonicalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func digestBytes(data []byte) string {
	h := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(h[:])
}

func NewSnapshotEnvelope(payload SnapshotPayload) (SnapshotEnvelope, error) {
	data, err := canonicalJSON(payload)
	if err != nil {
		return SnapshotEnvelope{}, err
	}
	return SnapshotEnvelope{
		Schema:        SnapshotEnvelopeSchema,
		PayloadSHA256: digestBytes(data),
		Payload:       payload,
	}, nil
}

func NewComparisonEnvelope(payload ComparisonPayload) (ComparisonEnvelope, error) {
	data, err := canonicalJSON(payload)
	if err != nil {
		return ComparisonEnvelope{}, err
	}
	return ComparisonEnvelope{
		Schema:        ComparisonEnvelopeSchema,
		PayloadSHA256: digestBytes(data),
		Payload:       payload,
	}, nil
}

func WriteCanonical(path string, value any) error {
	data, err := canonicalJSON(value)
	if err != nil {
		return err
	}
	if path == "-" {
		_, err = os.Stdout.Write(data)
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func ReadSnapshot(path string) (SnapshotEnvelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SnapshotEnvelope{}, err
	}
	var env SnapshotEnvelope
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&env); err != nil {
		return SnapshotEnvelope{}, fmt.Errorf("decode snapshot: %w", err)
	}
	if env.Schema != SnapshotEnvelopeSchema {
		return SnapshotEnvelope{}, fmt.Errorf("unsupported snapshot envelope schema %q", env.Schema)
	}
	if env.Payload.Schema != SnapshotSchema {
		return SnapshotEnvelope{}, fmt.Errorf("unsupported snapshot payload schema %q", env.Payload.Schema)
	}
	payload, err := canonicalJSON(env.Payload)
	if err != nil {
		return SnapshotEnvelope{}, err
	}
	if got := digestBytes(payload); got != env.PayloadSHA256 {
		return SnapshotEnvelope{}, fmt.Errorf("snapshot payload digest mismatch: got %s want %s", got, env.PayloadSHA256)
	}
	return env, nil
}

func VerifySnapshot(path string) error {
	_, err := ReadSnapshot(path)
	return err
}
