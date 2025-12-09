package game

import (
	"connect4/db"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	Username string
	Conn     *websocket.Conn
	IsBot    bool
}

type Game struct {
	ID        string
	P1, P2    *Player
	Board     Board
	Turn      int
	Active    bool
	StartTime time.Time
	Mutex     sync.Mutex
}

var MatchQueue = make(chan *Player, 100)

func StartMatchmaking() {
	for {
		p1 := <-MatchQueue
		log.Println(p1.Username, "joined queue. Waiting...")

		select {
		case p2 := <-MatchQueue:
			log.Println("Match found:", p1.Username, "vs", p2.Username)
			go StartGame(p1, p2)
		case <-time.After(10 * time.Second):
			log.Println("Timeout. Starting Bot game for", p1.Username)
			bot := &Player{Username: "Bot_AI", IsBot: true}
			go StartGame(p1, bot)
		}
	}
}

func StartGame(p1, p2 *Player) {
	game := &Game{
		P1: p1, P2: p2, Turn: 1, Active: true, StartTime: time.Now(),
	}

	broadcast(game, map[string]interface{}{"type": "START", "p1": p1.Username, "p2": p2.Username})

	go handleMoves(game, p1, 1)

	if !p2.IsBot {
		go handleMoves(game, p2, 2)
	} else {
		go func() {
			for game.Active {
				time.Sleep(1 * time.Second)
				game.Mutex.Lock()
				if game.Turn == 2 && game.Active {
					col := game.Board.BotMove()
					processMove(game, 2, col)
				}
				game.Mutex.Unlock()
			}
		}()
	}
}

func handleMoves(g *Game, p *Player, playerID int) {
	defer p.Conn.Close()
	for {
		_, msg, err := p.Conn.ReadMessage()
		if err != nil {
			g.Mutex.Lock()
			if g.Active {
				g.Active = false
				winner := g.P2.Username
				if playerID == 2 {
					winner = g.P1.Username
				}
				duration := time.Since(g.StartTime).Seconds()
				db.LogGameEnd(winner, duration)
				broadcast(g, map[string]interface{}{"type": "OVER", "winner": winner, "reason": "disconnect"})
			}
			g.Mutex.Unlock()
			return
		}

		var input map[string]int
		json.Unmarshal(msg, &input)
		col := input["col"]

		g.Mutex.Lock()
		if g.Turn == playerID && g.Active {
			processMove(g, playerID, col)
		}
		g.Mutex.Unlock()
	}
}

func processMove(g *Game, playerID int, col int) {
	if _, ok := g.Board.Drop(col, playerID); ok {
		if g.Board.CheckWin(playerID) {
			g.Active = false
			winnerName := g.P1.Username
			if playerID == 2 {
				winnerName = g.P2.Username
			}

			duration := time.Since(g.StartTime).Seconds()
			db.LogGameEnd(winnerName, duration)

			broadcast(g, map[string]interface{}{"type": "UPDATE", "board": g.Board, "turn": 0})
			broadcast(g, map[string]interface{}{"type": "OVER", "winner": winnerName})
		} else {
			g.Turn = 3 - playerID
			broadcast(g, map[string]interface{}{"type": "UPDATE", "board": g.Board, "turn": g.Turn})
		}
	}
}

func broadcast(g *Game, msg interface{}) {
	if !g.P1.IsBot && g.P1.Conn != nil {
		g.P1.Conn.WriteJSON(msg)
	}
	if !g.P2.IsBot && g.P2.Conn != nil {
		g.P2.Conn.WriteJSON(msg)
	}
}
