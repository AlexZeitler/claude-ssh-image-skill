package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	authToken string
	authOnce  sync.Once
	authError error
)

const tokenFileEnv = "CCIMGD_TOKEN_FILE"
const tokenFileDefault = ".ccimgd-token"

// getToken returns the shared auth token, loading it once on first call.
// Lookup order: CCIMGD_TOKEN_FILE env -> ~/.ccimgd-token
// If no file exists, one is generated and written.
func getToken() (string, error) {
	authOnce.Do(func() {
		path := os.Getenv(tokenFileEnv)
		if path == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				authError = fmt.Errorf("cannot determine home dir: %w", err)
				return
			}
			path = filepath.Join(home, tokenFileDefault)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				b := make([]byte, 32)
				if _, err := rand.Read(b); err != nil {
					authError = fmt.Errorf("failed to generate token: %w", err)
					return
				}
				token := hex.EncodeToString(b)
				if err := os.WriteFile(path, []byte(token), 0600); err != nil {
					authError = fmt.Errorf("failed to write token file %s: %w", path, err)
					return
				}
				authToken = token
				fmt.Fprintf(os.Stderr, "Generated new auth token in %s\n", path)
				return
			}
			authError = fmt.Errorf("failed to read token file %s: %w", path, err)
			return
		}

		authToken = strings.TrimSpace(string(data))
		if authToken == "" {
			authError = fmt.Errorf("token file %s is empty", path)
			return
		}
	})
	return authToken, authError
}

// hashToken returns the SHA-256 hash of the token for logging (never log the raw token).
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// validateToken checks that the provided token matches the shared secret using constant-time comparison.
func validateToken(provided string) bool {
	expected, err := getToken()
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}
