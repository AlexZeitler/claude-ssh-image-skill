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
	"syscall"
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

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			return
		}
		for _, b := range buf[:n] {
			if b == '\n' {
				goto ready
			}
		}
	}
ready:

	var resp response
	imgData, err := getClipboardImage()
	if err != nil {
		resp = response{OK: false, Error: err.Error()}
	} else {
		resp = response{OK: true, Image: base64.StdEncoding.EncodeToString(imgData)}
	}

	data, _ := json.Marshal(resp)
	data = append(data, '\n')
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
