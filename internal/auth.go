package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"time"
)

// https://www.melvinvivas.com/how-to-encrypt-and-decrypt-data-using-aes/

var gcm cipher.AEAD

func buildEmailVerificationToken(user User) (string, error) {
	// dump payload into buffer
	tokenPayload := EmailVerificationPayload{
		UserID:    user.ID,
		Email:     user.Email,
		Timestamp: time.Now().Unix(),
	}
	bytes, err := json.Marshal(tokenPayload)
	if err != nil {
		return "", err
	}

	// do encryption stuff
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	token := hex.EncodeToString(gcm.Seal(nonce, nonce, bytes, nil))
	return token, nil
}

func parseEmailVerificationToken(serializedToken string) (*EmailVerificationPayload, error) {
	// do decryption stuff
	token, err := hex.DecodeString(serializedToken)
	if err != nil {
		return nil, err
	}
	nonce, ciphertext := token[:gcm.NonceSize()], token[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)

	// unmarshall data
	payload := EmailVerificationPayload{}
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

func init() {
	secret := os.Getenv("EMAIL_VERIFICATION_SECRET")
	if secret == "" {
		secret = "dev_email_verification_secret___"
	}

	emailVerificationSecret := []byte(secret)

	block, err := aes.NewCipher(emailVerificationSecret)
	if err != nil {
		panic(err)
	}

	gcm, err = cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
}
