package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Settings for Argon2id hashing
const (
	// The amount of memory used by the algorithm (in kibibytes)
	memory = 64 * 1024
	// The number of iterations (passes) over the memory
	iterations = 3
	// The degree of parallelism (number of threads)
	parallelism = 2
	// The length of the salt (in bytes)
	saltLength = 16
	// The length of the hash (in bytes)
	keyLength = 32
)

// GenerateSalt generates a random salt of the specified length
func GenerateSalt() (string, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// Hash returns the Argon2id hash of the password using the provided salt
func Hash(password, salt string) (string, error) {
	saltBytes, err := base64.StdEncoding.DecodeString(salt)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), saltBytes, iterations, memory, parallelism, keyLength)
	return base64.StdEncoding.EncodeToString(hash), nil
}

// Verify checks if the provided password matches the stored hash using the salt
func Verify(password, storedHash, salt string) bool {
	hash, err := Hash(password, salt)
	if err != nil {
		return false
	}

	// Use a constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(hash), []byte(storedHash)) == 1
}

// GenerateHashAndSalt generates a salt and hash for the given password
func GenerateHashAndSalt(password string) (hash string, salt string, err error) {
	salt, err = GenerateSalt()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash, err = Hash(password, salt)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	return hash, salt, nil
}
