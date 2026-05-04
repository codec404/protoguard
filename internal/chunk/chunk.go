package chunk

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codec404/protoguard/internal/model"
)

// Chunk is a bounded set of changes for one LLM request.
type Chunk struct {
	ID      string         `json:"id"`
	Changes []model.Change `json:"changes"`
}

const DefaultMaxChunkBytes = 8000

// Split groups changes by openapi path or grpc service, merging small groups until maxBytes.
func Split(report *model.DiffReport, maxBytes int) ([]Chunk, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxChunkBytes
	}
	buckets := map[string][]model.Change{}
	order := []string{}

	add := func(key string, ch model.Change) {
		if _, ok := buckets[key]; !ok {
			order = append(order, key)
		}
		buckets[key] = append(buckets[key], ch)
	}

	for _, ch := range report.Changes {
		key := bucketKey(ch.Path)
		add(key, ch)
	}

	var out []Chunk
	for _, k := range order {
		parts := splitToMaxBytes(k, buckets[k], maxBytes)
		out = append(out, parts...)
	}
	if len(out) == 0 {
		out = append(out, Chunk{ID: "empty", Changes: nil})
	}
	return out, nil
}

func bucketKey(path string) string {
	if strings.HasPrefix(path, "openapi.paths.") {
		rest := strings.TrimPrefix(path, "openapi.paths.")
		idx := strings.Index(rest, ".")
		if idx <= 0 {
			return "openapi.paths." + rest
		}
		return "openapi.paths." + rest[:idx]
	}
	if strings.HasPrefix(path, "grpc.service.") {
		rest := strings.TrimPrefix(path, "grpc.service.")
		dot := strings.Index(rest, ".method.")
		if dot > 0 {
			return "grpc.service." + rest[:dot]
		}
		return "grpc.service." + rest
	}
	if strings.HasPrefix(path, "grpc.field.") {
		rest := strings.TrimPrefix(path, "grpc.field.")
		dot := strings.Index(rest, ".")
		if dot > 0 {
			return "grpc.field." + rest[:dot]
		}
	}
	if strings.HasPrefix(path, "openapi.components.") {
		return "openapi.components"
	}
	return "misc"
}

func splitToMaxBytes(id string, changes []model.Change, maxBytes int) []Chunk {
	var res []Chunk
	var cur []model.Change
	curID := id + "-0"
	n := 0

	flush := func() {
		if len(cur) == 0 {
			return
		}
		res = append(res, Chunk{ID: curID, Changes: append([]model.Change(nil), cur...)})
		cur = cur[:0]
		n++
		curID = fmt.Sprintf("%s-%d", id, n)
	}

	for _, ch := range changes {
		trial := append(cur, ch)
		b, err := json.Marshal(trial)
		if err != nil {
			return []Chunk{{ID: id, Changes: changes}}
		}
		if len(b) > maxBytes && len(cur) > 0 {
			flush()
			trial = []model.Change{ch}
			b, _ = json.Marshal(trial)
		}
		if len(b) > maxBytes {
			res = append(res, Chunk{ID: curID + "-singleton", Changes: []model.Change{ch}})
			n++
			curID = fmt.Sprintf("%s-%d", id, n)
			continue
		}
		cur = trial
	}
	flush()
	return res
}
