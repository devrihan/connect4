package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type GameEvent struct {
	Event     string    `json:"event"`
	Winner    string    `json:"winner"`
	Duration  float64   `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

type Metrics struct {
	TotalGames    int            `json:"totalGames"`
	TotalDuration float64        `json:"-"`
	AvgDuration   float64        `json:"avgDuration"`
	WinsPerUser   map[string]int `json:"winsPerUser"`
	GamesPerHour  map[int]int    `json:"gamesPerHour"`
	Mutex         sync.RWMutex   `json:"-"`
}

var stats = Metrics{
	WinsPerUser:  make(map[string]int),
	GamesPerHour: make(map[int]int),
}

func main() {
	go startHTTPServer()

	startKafkaConsumer()
}

func startHTTPServer() {
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		stats.Mutex.RLock()
		defer stats.Mutex.RUnlock()
		json.NewEncoder(w).Encode(stats)
	})

	log.Println("Analytics API running on http://localhost:8081/stats")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func startKafkaConsumer() {
	topic := "game-analytics"
	partition := 0

	log.Println("Connecting to Kafka...")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     topic,
		Partition: partition,
		MinBytes:  10e3,
		MaxBytes:  10e6,
	})
	defer r.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down Analytics Service")
		r.Close()
		os.Exit(0)
	}()

	log.Println("Listening for Game Events...")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var event GameEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Error parsing JSON: %v\n", err)
			continue
		}

		processEvent(event)
	}
}

func processEvent(e GameEvent) {
	stats.Mutex.Lock()
	defer stats.Mutex.Unlock()

	stats.TotalGames++
	stats.TotalDuration += e.Duration

	if stats.TotalGames > 0 {
		stats.AvgDuration = stats.TotalDuration / float64(stats.TotalGames)
	}

	if e.Winner != "" {
		stats.WinsPerUser[e.Winner]++
	}

	hour := e.Timestamp.Hour()
	stats.GamesPerHour[hour]++

	fmt.Printf("Updated Stats: %d Games, Avg Duration: %.2fs\n", stats.TotalGames, stats.AvgDuration)
}
