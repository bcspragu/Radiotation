// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[string]*connection

	// Inbound messages from the connections.
	broadcast chan []byte

	// Register requests from the connections.
	register chan *connection

	// Unregister requests from connections.
	unregister chan *connection
}

var h = hub{
	broadcast:   make(chan []byte),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
	connections: make(map[string]*connection),
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c.loginID] = c
		case c := <-h.unregister:
			if _, ok := h.connections[c.loginID]; ok {
				delete(h.connections, c.loginID)
				close(c.send)
			}
		case m := <-h.broadcast:
			for _, c := range h.connections {
				select {
				case c.send <- m:
				default:
					close(c.send)
					delete(h.connections, c.loginID)
				}
			}
		}
	}
}
