package db

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var DB *sql.DB
var KafkaWriter *kafka.Writer
var KafkaEnabled bool = false

func InitDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "host=localhost user=user password=password dbname=connect4 sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(" Failed to open DB driver:", err)
	}

	for i := 0; i < 10; i++ {
		if err = DB.Ping(); err == nil {
			break
		}
		log.Println(" Waiting for Database to start...", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("❌ Could not connect to Database after retries:", err)
	}
	log.Println("✅ Connected to Database")

	query := `
	CREATE TABLE IF NOT EXISTS users (username VARCHAR(50) PRIMARY KEY, wins INT DEFAULT 0);
	CREATE TABLE IF NOT EXISTS games (id SERIAL PRIMARY KEY, winner VARCHAR(50), timestamp TIMESTAMP);
	`
	_, err = DB.Exec(query)
	if err != nil {
		log.Fatal("❌ Failed to create tables:", err)
	}
}

func InitKafka() {
	broker := os.Getenv("KAFKA_BROKER")

	if broker == "" {
		log.Println(" No KAFKA_BROKER set. Analytics will be disabled.")
		KafkaEnabled = false
		return
	}

	username := os.Getenv("KAFKA_USERNAME")
	password := os.Getenv("KAFKA_PASSWORD")

	writerConfig := kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   "game-analytics",
		Dialer: &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		},
	}

	if username != "" && password != "" {
		writerConfig.Dialer.SASLMechanism = plain.Mechanism{
			Username: username,
			Password: password,
		}
		writerConfig.Dialer.TLS = &tls.Config{}
	}

	KafkaWriter = kafka.NewWriter(writerConfig)
	KafkaEnabled = true
	log.Println("✅ Kafka Analytics Enabled")
}

func LogGameEnd(winner string, duration float64) {
	if winner != "Draw" {
		_, err := DB.Exec("INSERT INTO users (username, wins) VALUES ($1, 1) ON CONFLICT (username) DO UPDATE SET wins = users.wins + 1", winner)
		if err != nil {
			log.Println("DB Error:", err)
		}
	}
	DB.Exec("INSERT INTO games (winner, timestamp) VALUES ($1, $2)", winner, time.Now())

	if KafkaEnabled && KafkaWriter != nil {
		go func() {
			event := map[string]interface{}{
				"event":     "GAME_OVER",
				"winner":    winner,
				"duration":  duration,
				"timestamp": time.Now(),
			}
			msg, _ := json.Marshal(event)

			err := KafkaWriter.WriteMessages(context.Background(),
				kafka.Message{Value: msg},
			)
			if err != nil {
				log.Println("Kafka Write Error:", err)
			} else {
				log.Printf("Analytics sent: Winner=%s Duration=%.2fs\n", winner, duration)
			}
		}()
	}
}

func GetLeaderboard() []map[string]interface{} {
	rows, err := DB.Query("SELECT username, wins FROM users ORDER BY wins DESC LIMIT 10")
	if err != nil {
		log.Println("Leaderboard Query Error:", err)
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
