package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// Dir returns default cache directory (XDG_CACHE_HOME or ~/.cache).
func Dir(custom string) (string, error) {
	if custom != "" {
		return custom, os.MkdirAll(custom, 0o700)
	}
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".cache")
	}
	dir := filepath.Join(base, "protoguard")
	return dir, os.MkdirAll(dir, 0o700)
}

// Key builds a cache key from parts (normalized JSON payload + metadata).
func Key(parts ...[]byte) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write(p)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Path joins cache dir with key hex prefix subdirs.
func Path(dir, key string) string {
	if len(key) < 4 {
		return filepath.Join(dir, key+".json")
	}
	return filepath.Join(dir, key[:2], key+".json")
}

// Get reads cached JSON bytes if present.
func Get(dir, key string) ([]byte, bool, error) {
	p := Path(dir, key)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return b, true, nil
}

// Put writes JSON bytes to cache path atomically (0600 — may contain sensitive API metadata).
func Put(dir, key string, data []byte) error {
	p := Path(dir, key)
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

// MarshalStable JSON for hashing payloads.
func MarshalStable(v any) ([]byte, error) {
	return json.Marshal(v)
}
