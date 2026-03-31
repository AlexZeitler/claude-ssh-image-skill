package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	host = "127.0.0.1"
	port = "9998"
)

type response struct {
	OK    bool   `json:"ok"`
	Image string `json:"image,omitempty"`
	Error string `json:"error,omitempty"`
}

func getClipboardImage() ([]byte, error) {
	var cmd *exec.Cmd
	switch {
	case runtime.GOOS == "darwin":
		cmd = exec.Command("pngpaste", "-")
	case os.Getenv("WAYLAND_DISPLAY") != "":
		cmd = exec.Command("wl-paste", "--type", "image/png")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard", "-target", "image/png", "-o")
	}
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return nil, fmt.Errorf("Clipboard is empty or does not contain an image")
	}
	return out, nil
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	remoteAddr := conn.RemoteAddr().String()
	buf := make([]byte, 4096)
	totalRead := 0

	for totalRead < len(buf) {
		n, err := conn.Read(buf[totalRead:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Connection from %s: read error: %v\n", remoteAddr, err)
			return
		}
		totalRead += n
		for _, b := range buf[:totalRead] {
			if b == '\n' {
				goto ready
			}
		}
	}
	fmt.Fprintf(os.Stderr, "Connection from %s: buffer overflow, dropping\n", remoteAddr)
	return

ready:
	request := strings.TrimSpace(string(buf[:totalRead]))

	var req map[string]string
	json.Unmarshal([]byte(request), &req)
	reqToken := req["token"]

	if !validateToken(reqToken) {
		fmt.Fprintf(os.Stderr, "Connection from %s: auth failed (token hash: %s)\n", remoteAddr, hashToken(reqToken))
		resp := response{OK: false, Error: "unauthorized"}
		data, _ := json.Marshal(resp)
		data = append(data, '\n')
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.Write(data)
		return
	}

	fmt.Fprintf(os.Stderr, "Connection from %s: authenticated (token hash: %s)\n", remoteAddr, hashToken(reqToken))

	var resp response
	imgData, err := getClipboardImage()
	if err != nil {
		resp = response{OK: false, Error: err.Error()}
	} else {
		resp = response{OK: true, Image: base64.StdEncoding.EncodeToString(imgData)}
	}

	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	conn.Write(data)
}

func main() {
	listener, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("ccimgd listening on %s:%s\n", host, port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		handleConn(conn)
	}
}
