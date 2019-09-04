package main

import (
	"fmt"
	"log"
	"math/rand"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	turn       chan *Client
	playedTurn chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		turn:       make(chan *Client),
		playedTurn: make(chan *Client),
	}
}

func (h *Hub) startGame() error {
	if len(h.clients) == 0 {
		return fmt.Errorf("no player")
	}
	// Reset all turns
	for client := range h.clients {
		h.clients[client] = false
	}
	h.selectTurn() // Start first turn

	return nil
}

func (h *Hub) selectTurn() {
	notYetPlayed := make([]*Client, 0)
	for client, played := range h.clients {
		if !played && client.name != "admin" {
			notYetPlayed = append(notYetPlayed, client)
		}
	}

	turn := rand.Intn(len(notYetPlayed))
	h.turn <- notYetPlayed[turn]
	h.clients[notYetPlayed[turn]] = true
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = false
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case client := <-h.turn:
			select {
			case client.send <- []byte(fmt.Sprintf("turn:%s", client.name)):
			default:
				close(client.send)
				delete(h.clients, client)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				log.Println("Sending to client:", client.name)
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
