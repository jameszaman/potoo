package sendgrid

import "crypto/sha256"

func hashPayload(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
