package gccmp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return SnapshotEnvelope{}, fmt.Errorf("decode snapshot: %w", err)
	}

	var env SnapshotEnvelope
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&env); err != nil {
		return SnapshotEnvelope{}, fmt.Errorf("decode snapshot: %w", err)
	}
	if err := requireJSONEOF(dec); err != nil {
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

// rejectDuplicateJSONKeys performs a syntax walk before typed decoding. Go's
// normal object decoder accepts duplicate member names with last-value-wins
// behavior; canonical evidence cannot permit that ambiguity.
func rejectDuplicateJSONKeys(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := walkJSONValue(dec); err != nil {
		return err
	}
	return requireJSONEOF(dec)
}

func walkJSONValue(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object member name is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate JSON object key %q", key)
			}
			seen[key] = struct{}{}
			if err := walkJSONValue(dec); err != nil {
				return err
			}
		}
		end, err := dec.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return fmt.Errorf("malformed JSON object")
		}
	case '[':
		for dec.More() {
			if err := walkJSONValue(dec); err != nil {
				return err
			}
		}
		end, err := dec.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return fmt.Errorf("malformed JSON array")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	return nil
}

func requireJSONEOF(dec *json.Decoder) error {
	var trailing any
	err := dec.Decode(&trailing)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("trailing JSON value after snapshot envelope")
	}
	return fmt.Errorf("trailing non-whitespace data after snapshot envelope: %w", err)
}
