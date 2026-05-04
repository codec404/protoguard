package security

import (
	"fmt"
	"io"
	"os"
)

// ReadSpecFile reads path entirely into memory, rejecting files larger than maxBytes.
func ReadSpecFile(path string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxSpecBytes
	}
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.Size() > maxBytes {
		return nil, fmt.Errorf("spec file %q too large (%d bytes; max %d)", path, fi.Size(), maxBytes)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	limited := io.LimitReader(f, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("spec file %q exceeds max size %d bytes", path, maxBytes)
	}
	return data, nil
}
