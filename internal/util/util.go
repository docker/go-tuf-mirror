package util

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
)

func HexHashBytes(input []byte) string {
	s256 := sha256.New()
	s256.Write(input)
	hashSum := s256.Sum(nil)
	return hex.EncodeToString(hashSum)
}

// isValidUrl tests a string to determine if it is a well-structured url or not.
func IsValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	return err == nil && u.Scheme != "" && u.Host != ""
}
