package securityhelper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Mask access token, using asterisk character for 4 digits in the beginning and 4 digits in the last
func MaskToken(token string) string {
	if len(token) < 12 {
		return "********"
	}
	return fmt.Sprintf("%s....%s", token[:6], token[len(token)-6:])
}

// Hash access token using encryption SHA 256
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:8])
}
