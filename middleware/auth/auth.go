package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func IsAuthorized(authHeader string, body []byte) bool {

	return EncryptHeader(body) == authHeader
}

func EncryptHeader(body []byte) string {
	hash := hmac.New(sha256.New, []byte(os.Getenv("SECRET")))
	hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil))
}
