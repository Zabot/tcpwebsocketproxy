package tcpwebsocketproxy

import (
	"net"

	"github.com/gorilla/websocket"
)

func proxyPass(ws_conn *websocket.Conn, ip_conn net.Conn) {
	ws_rx := make(chan []byte)
	go func() {
		for {
			_, bytes, _ := ws_conn.ReadMessage()
			if len(bytes) > 0 {
				ws_rx <- bytes
			}
		}
	}()

	ip_rx := make(chan []byte)
	go func() {
		buf := make([]byte, 256)
		for {
			n, _ := ip_conn.Read(buf)
			if n > 0 {
				ip_rx <- buf[0:n]
			}
		}
	}()

	for {
		select {
		case tcp_message := <-ip_rx:
			ws_conn.WriteMessage(websocket.BinaryMessage, tcp_message)

		case ws_message := <-ws_rx:
			ip_conn.Write(ws_message)
		}
	}
}
