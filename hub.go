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

func newHub() Hub {
	return Hub{
		Broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

var Rooms map[string]map[*Client]bool = make(map[string]map[*Client]bool)

func (h Hub) run() {
	for {
		fmt.Println(Rooms)
		select {
		case client := <-h.register:
			roomTest := Rooms[client.roomID]
			if roomTest == nil {
				r := make(map[*Client]bool)
				Rooms[client.roomID] = r
			}
			Rooms[client.roomID][client] = true
			break
			// TEST
			r := make(map[*Client]bool)
			Rooms[client.roomID] = r
			r[client] = true
			break
			room := h.rooms[client.roomID]
			rr := Rooms[client.roomID]
			if rr == nil {
				fmt.Println("--> create new room")
				rr = make(map[*Client]bool)
				Rooms = make(map[string]map[*Client]bool)
				Rooms[client.roomID] = rr
			}
			// TEST

			if room == nil {
				// fmt.Println("--> create new room")
				// First client in the room, create a new one
				room = make(map[*Client]bool)
				h.rooms = make(map[string]map[*Client]bool)
				h.rooms[client.roomID] = room

				Rooms = make(map[string]map[*Client]bool)
				Rooms[client.roomID] = room
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
			room := Rooms[message.roomID]
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
