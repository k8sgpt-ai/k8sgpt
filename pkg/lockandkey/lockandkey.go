package lockandkey

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func Encrypt(key, plaintext []byte) (string, error) {
	// Create a new cipher block using the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a new GCM block cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the plaintext using GCM
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return string(ciphertext), nil
}

func Decrypt(key, ciphertext []byte) (string, error) {
	// Create a new cipher block using the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a new GCM block cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Get the nonce size from the cipher
	nonceSize := gcm.NonceSize()

	// Verify the size of the ciphertext
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract the nonce from the ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the ciphertext using GCM
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
