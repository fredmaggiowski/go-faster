package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	clients map[*Client]bool

	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	turn       chan *Client
	playedTurn chan *Client
	statistics *Statistics
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 1),
		turn:       make(chan *Client, 1),
		playedTurn: make(chan *Client, 1),
		statistics: newStatistics(),
	}
}

func (h *Hub) startGame() error {
	if len(h.clients) == 0 {
		return fmt.Errorf("no player")
	}

	for client := range h.clients {
		h.clients[client] = false
	}
	h.statistics = newStatistics()
	h.broadcast <- []byte("game:start")

	if !h.selectTurn() {
		return fmt.Errorf("game over")
	}

	return nil
}

func (h *Hub) selectTurn() bool {
	notYetPlayed := make([]*Client, 0)
	for client, played := range h.clients {
		if !played && client.name != "admin" {
			notYetPlayed = append(notYetPlayed, client)
		}
	}
	if len(notYetPlayed) == 0 {
		return false
	}

	turn := rand.Intn(len(notYetPlayed))
	h.turn <- notYetPlayed[turn]
	h.clients[notYetPlayed[turn]] = true
	return true
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
				h.statistics.Speed[client.name] = TimeTrack{Start: time.Now()}
			default:
				close(client.send)
				delete(h.clients, client)
			}

		case client := <-h.playedTurn:
			h.clients[client] = true
			playerTrack := h.statistics.Speed[client.name]
			playerTrack.End = time.Now()
			h.statistics.Speed[client.name] = playerTrack

			delta := playerTrack.End.Sub(playerTrack.Start).Seconds()
			log.Printf("Player %s took %f.03s\n", client.name, delta)
			if !h.selectTurn() {
				log.Println("No more players, game is over")
				h.broadcast <- []byte(fmt.Sprintf("game:over\nwinner:%s", h.statistics.getWinner()))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				log.Printf("Broadcasting message to client %s.\n", client.name)
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
