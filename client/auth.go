package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const tokenFileEnv = "CCIMGD_TOKEN_FILE"
const tokenFileDefault = ".ccimgd-token"

func loadToken() (string, error) {
	path := os.Getenv(tokenFileEnv)
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home dir: %w", err)
		}
		path = filepath.Join(home, tokenFileDefault)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read token file %s: %w", path, err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file %s is empty", path)
	}
	return token, nil
}
