package hub

import (
	"github.com/gorilla/websocket"
)

// Hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	connections map[*connection]bool

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

func New() *Hub {
	h := &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[*connection]bool),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			if _, ok := h.connections[c]; ok {
				delete(h.connections, c)
				close(c.send)
			}
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(msg []byte) {
	h.broadcast <- msg
}

// Register associates a connection with the hub.
func (h *Hub) Register(ws *websocket.Conn) {
	conn := &connection{h: h, send: make(chan []byte, 256), ws: ws}
	h.register <- conn
	go conn.writePump()
	go conn.readPump()
}
