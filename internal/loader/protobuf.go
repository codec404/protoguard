package loader

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	dpb "google.golang.org/protobuf/types/descriptorpb"

	"github.com/codec404/protoguard/internal/security"
)

// LoadFileDescriptorSet reads a protobuf-encoded google.protobuf.FileDescriptorSet from a local path.
func LoadFileDescriptorSet(path string, opts Options) (*dpb.FileDescriptorSet, error) {
	max := opts.MaxBytes
	if max <= 0 {
		max = security.DefaultMaxSpecBytes
	}
	data, err := security.ReadSpecFile(path, max)
	if err != nil {
		return nil, err
	}
	var fds dpb.FileDescriptorSet
	if err := proto.Unmarshal(data, &fds); err != nil {
		return nil, fmt.Errorf("decode FileDescriptorSet %s: %w", path, err)
	}
	if len(fds.File) == 0 {
		return nil, fmt.Errorf("empty FileDescriptorSet %s", path)
	}
	return &fds, nil
}
