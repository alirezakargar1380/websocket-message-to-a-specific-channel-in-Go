// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type Message struct {
	roomID string
	Data   []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool
	rooms   map[string]map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// var roomss Hub.rooms
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:

			room := h.rooms[client.roomID]
			if room == nil {
				fmt.Println("--> create new room")
				// First client in the room, create a new one
				room = make(map[*Client]bool)
				h.rooms = make(map[string]map[*Client]bool)
				h.rooms[client.roomID] = room
			}
			room[client] = true
		case client := <-h.unregister:
			room := h.rooms[client.roomID]
			if room != nil {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.send)
					if len(room) == 0 {
						// This was last client in the room, delete the room
						delete(h.rooms, client.roomID)
					}
				}
			}
		case message := <-h.Broadcast:
			// fmt.Println(message.Data)
			room := h.rooms[message.roomID]
			fmt.Println("message room ID: " + message.roomID)
			fmt.Println(room)
			if room != nil {
				for client := range room {
					select {
					case client.send <- message.Data:
					default:
						close(client.send)
						delete(room, client)
					}
				}
				if len(room) == 0 {
					// The room was emptied while broadcasting to the room.  Delete the room.
					delete(h.rooms, message.roomID)
				}
			}
		}

	}
}
