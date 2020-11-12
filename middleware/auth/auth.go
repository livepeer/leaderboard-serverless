package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
)

// 403 Forbidden

func IsAuthorized(authHeader string, body []byte) bool {
	hash := hmac.New(sha256.New, []byte(os.Getenv("SECRET")))
	hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil)) == authHeader
}
