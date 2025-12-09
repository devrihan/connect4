package main

import (
	"connect4/db"
	"connect4/game"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	player := &game.Player{Username: username, Conn: conn}
	game.MatchQueue <- player
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	data := db.GetLeaderboard()
	json.NewEncoder(w).Encode(data)
}

func main() {
	time.Sleep(5 * time.Second)
	db.InitDB()
	db.InitKafka()

	go game.StartMatchmaking()

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/leaderboard", leaderboardHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
