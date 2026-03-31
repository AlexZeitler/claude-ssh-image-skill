package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const (
	host = "127.0.0.1"
	port = "9998"
)

type response struct {
	OK    bool   `json:"ok"`
	Image string `json:"image"`
	Error string `json:"error"`
}

func main() {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to ccimgd: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	token, err := loadToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load auth token: %v\n", err)
		os.Exit(1)
	}

	req := map[string]string{"token": token}
	reqData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal request: %v\n", err)
		os.Exit(1)
	}
	reqData = append(reqData, '\n')

	_, err = conn.Write(reqData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send request: %v\n", err)
		os.Exit(1)
	}

	var buf []byte
	tmp := make([]byte, 65536)
	for {
		n, err := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
		for _, b := range tmp[:n] {
			if b == '\n' {
				goto done
			}
		}
	}
done:

	var resp response
	if err := json.Unmarshal(buf, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if !resp.OK {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		os.Exit(1)
	}

	imgData, err := base64.StdEncoding.DecodeString(resp.Image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode image: %v\n", err)
		os.Exit(1)
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("clipboard-%d.png", os.Getpid()))
	if err := os.WriteFile(tmpFile, imgData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write image: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(tmpFile)
}
