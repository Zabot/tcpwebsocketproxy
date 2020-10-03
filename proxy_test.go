package tcpwebsocketproxy

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestLoopback(t *testing.T) {
	// Start an echo server running on localhost:4000
	ln, err := net.Listen("tcp", "127.0.0.1:4000")
	if err != nil {
		t.Fatalf("[SERVER] Failed to start echo server: %v", err)
	}
	t.Logf("[SERVER] Running echo server on 127.0.0.1:4000")

	go func() {
		server_conn, err := ln.Accept()
		if err != nil {
			t.Fatalf("[SERVER] Failed to accept connection: %v", err)
		}
		defer server_conn.Close()
		t.Logf("[SERVER] Accepted connection from %v", server_conn.RemoteAddr())

		buf := make([]byte, 256)
		for {
			n, err := server_conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("[SERVER] Read failed: %v", err)
			}
			if n > 0 {
				t.Logf("[SERVER] Read %d bytes", n)
				n, err := server_conn.Write(buf[0:n])
				if err != nil {
					t.Fatalf("[SERVER] Write failed: %v", err)
				}
				t.Logf("[SERVER] Wrote %d bytes", n)
			}
		}
		t.Logf("[SERVER] Connection closed")
	}()

	// Proxy localhost:3000 to localhost:4000 over websockets
	go ProxyWebsocket("/proxy", "localhost:4000")
	go ProxyPort(3000, "ws://localhost:8080/proxy")

	// Connect to the the proxied server
	var client_conn net.Conn
	start := time.Now()
	for {
		client_conn, err = net.Dial("tcp", "127.0.0.1:3000")
		if err == nil {
			break
		}
		if time.Since(start) > 50*time.Millisecond {
			t.Fatalf("[CLIENT] Timed out trying to connect")
		} else if err != nil {
			t.Logf("[CLIENT] Failed to connect to proxied port %v", err)
		}
	}
	defer client_conn.Close()
	t.Logf("[CLIENT] Connected to echo server")

	// Send data to the server
	test := func(s string) {
		n, err := client_conn.Write([]byte(s))
		if err != nil {
			t.Fatalf("[CLIENT] Failed to write: %v", err)
		}
		t.Logf("[CLIENT] Wrote %v bytes", n)

		buf := make([]byte, 256)

		n, err = client_conn.Read(buf)
		if err != nil {
			t.Fatalf("[CLIENT] Failed to read: %v", err)
		}

		t.Logf("[CLIENT] Read %d bytes", n)

		if string(buf[0:n]) != s {
			t.Fatalf("[CLIENT] Expected %v but got %v", s, string(buf))
		}
	}

	test("Hello world\n")
}
