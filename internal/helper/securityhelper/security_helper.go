package securityhelper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func MaskToken(token string) string {
	if len(token) < 12 {
		return "********"
	}
	return fmt.Sprintf("%s....%s", token[:6], token[len(token)-6:])
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:8])
}
