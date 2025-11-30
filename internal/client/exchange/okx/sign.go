package okx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func AccessSign(timestamp, method, requestPath, body, secretKey string) string {
	payload := timestamp + method + requestPath + body
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))
	signature := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(signature)
}
