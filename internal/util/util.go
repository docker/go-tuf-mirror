package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func HexHashBytes(input []byte) string {
	s256 := sha256.New()
	s256.Write(input)
	hashSum := s256.Sum(nil)
	return hex.EncodeToString(hashSum)
}
