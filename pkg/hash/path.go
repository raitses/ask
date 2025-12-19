package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// DirectoryPath computes a short hash of an absolute directory path
// for use as a context file identifier.
func DirectoryPath(path string) string {
	h := sha256.New()
	h.Write([]byte(path))
	return hex.EncodeToString(h.Sum(nil))[:8]
}
