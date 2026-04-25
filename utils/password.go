package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
)

const passwordRounds = 120000

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := makePasswordHash(password, salt)
	return "sha256$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(hash), nil
}

func CheckPassword(password string, stored string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 3 || parts[0] != "sha256" {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	expected, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	actual := makePasswordHash(password, salt)
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func makePasswordHash(password string, salt []byte) []byte {
	input := append([]byte{}, salt...)
	input = append(input, []byte(password)...)
	sum := sha256.Sum256(input)
	hash := sum[:]

	for i := 0; i < passwordRounds; i++ {
		next := append([]byte{}, hash...)
		next = append(next, salt...)
		next = append(next, []byte(password)...)
		sum = sha256.Sum256(next)
		hash = sum[:]
	}

	return append([]byte{}, hash...)
}

func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must contain at least 6 characters")
	}
	return nil
}
