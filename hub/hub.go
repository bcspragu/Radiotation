package hub

import (
	"fmt"
	"math/rand"

	"github.com/bcspragu/Radiotation/db"
	"github.com/gorilla/websocket"
)

// Hub maintains the set of active connections and broadcasts messages to the
// connections.
type Hub struct {
	// Registered connections.
	connections map[db.RoomID][]*connection

	// Inbound messages from the connections.
	broadcast chan *broadcastMsg

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

// New creates a new Hub and starts it in a background Go routine.
func New() *Hub {
	h := &Hub{
		broadcast:   make(chan *broadcastMsg),
		register:    make(chan *connection),
		unregister:  make(chan *connection),
		connections: make(map[db.RoomID][]*connection),
	}
	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			conns := h.connections[c.rm.ID]
			h.connections[c.rm.ID] = append(conns, c)
		case c := <-h.unregister:
			h.deleteConn(c)
		case m := <-h.broadcast:
			for _, c := range h.connections[m.roomID] {
				select {
				case c.send <- m.msg:
				default:
					h.deleteConn(c)
				}
			}
		}
	}
}

func (h *Hub) deleteConn(c *connection) {
	close(c.send)
	rconns := h.connections[c.rm.ID]
	for i, rconn := range rconns {
		if rconn.id == c.id {
			// Remove the connection.
			copy(rconns[i:], rconns[i+1:])
			rconns[len(rconns)-1] = nil
			h.connections[c.rm.ID] = rconns[:len(rconns)-1]
			return
		}
	}
}

type broadcastMsg struct {
	roomID db.RoomID
	msg    []byte
}

// BroadcastRoom sends a message to everyone in a room.
func (h *Hub) BroadcastRoom(msg []byte, rm *db.Room) {
	h.broadcast <- &broadcastMsg{
		roomID: rm.ID,
		msg:    msg,
	}
}

// Register associates a connection with the hub and a given room.
func (h *Hub) Register(ws *websocket.Conn, rm *db.Room) {
	conn := &connection{id: newID(rm), h: h, rm: rm, send: make(chan []byte, 256), ws: ws}
	h.register <- conn
	go conn.writePump()
	go conn.readPump()
}

func newID(rm *db.Room) string {
	return fmt.Sprintf("%s-%d", rm.ID, rand.Int63())
}
