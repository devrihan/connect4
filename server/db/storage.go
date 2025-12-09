package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

var DB *sql.DB
var KafkaWriter *kafka.Writer

func InitDB() {
	connStr := "user=user password=password dbname=connect4 sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Create tables if not exist
	query := `
	CREATE TABLE IF NOT EXISTS users (username VARCHAR(50) PRIMARY KEY, wins INT DEFAULT 0);
	CREATE TABLE IF NOT EXISTS games (id SERIAL PRIMARY KEY, winner VARCHAR(50), timestamp TIMESTAMP);
	`
	_, err = DB.Exec(query)
	if err != nil {
		log.Println("Waiting for DB...", err) // Simple retry logic might be needed here in prod
	}
}

func InitKafka() {
	KafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "game-analytics",
		Balancer: &kafka.LeastBytes{},
	}
}

func LogGameEnd(winner string) {
	// 1. Update DB
	if winner != "Draw" {
		_, err := DB.Exec("INSERT INTO users (username, wins) VALUES ($1, 1) ON CONFLICT (username) DO UPDATE SET wins = users.wins + 1", winner)
		if err != nil {
			log.Println("DB Error:", err)
		}
	}
	DB.Exec("INSERT INTO games (winner, timestamp) VALUES ($1, $2)", winner, time.Now())

	// 2. Send to Kafka
	event := map[string]interface{}{
		"event":     "GAME_OVER",
		"winner":    winner,
		"timestamp": time.Now(),
	}
	msg, _ := json.Marshal(event)

	err := KafkaWriter.WriteMessages(context.Background(),
		kafka.Message{Value: msg},
	)
	if err != nil {
		log.Println("Kafka Error:", err)
	}
}

func GetLeaderboard() []map[string]interface{} {
	rows, err := DB.Query("SELECT username, wins FROM users ORDER BY wins DESC LIMIT 10")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var u string
		var w int
		rows.Scan(&u, &w)
		results = append(results, map[string]interface{}{"username": u, "wins": w})
	}
	return results
}
