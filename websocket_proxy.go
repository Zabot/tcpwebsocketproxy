package tcpwebsocketproxy

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
)

/// Listen on a local port port and proxy connections to a remote websocket
func ProxyPort(local_port uint16, ws_url string) {
	addr := fmt.Sprintf("127.0.0.1:%d", local_port)
	log.Printf("Proxying %v to %v", addr, ws_url)

	ln, _ := net.Listen("tcp", addr)
	for {
		ip_conn, _ := ln.Accept()
		go func(ip_conn net.Conn, u string) {
			// Open a connection to the websocket
			ws_conn, _, _ := websocket.DefaultDialer.Dial(u, nil)
			defer ws_conn.Close()

			// Proxy
			proxyPass(ws_conn, ip_conn)
		}(ip_conn, ws_url)
	}
}

/// Listen on a local websocket and proxy connections to a remote port
func ProxyWebsocket(route string, addr string) {
	log.Printf("Proxying %v to %v", route, addr)

	upgrader := websocket.Upgrader{}
	proxy := func(w http.ResponseWriter, r *http.Request) {
		// Accept the websocket connection
		ws_conn, _ := upgrader.Upgrade(w, r, nil)
		defer ws_conn.Close()

		// Open the target IP socket
		ip_conn, _ := net.Dial("tcp", addr)
		defer ip_conn.Close()

		// Proxy
		proxyPass(ws_conn, ip_conn)
	}

	http.HandleFunc(route, proxy)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
