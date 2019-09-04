package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func getWSConnection(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, hub.unregister, hub.broadcast, hub.playedTurn)
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func main() {
	hub := newHub()
	go hub.run()

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, r, "game.html")
	})
	router.HandleFunc("/old", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, r, "home.html")
	})
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		getWSConnection(hub, w, r)
	})
	router.HandleFunc("/startgame", func(w http.ResponseWriter, r *http.Request) {
		if err := hub.startGame(); err != nil {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	})
	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func serveFile(w http.ResponseWriter, r *http.Request, file string) {
	log.Println(r.URL)
	// if r.URL.Path != "/" {
	// 	http.Error(w, "Not found", http.StatusNotFound)
	// 	return
	// }
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, file)
}
