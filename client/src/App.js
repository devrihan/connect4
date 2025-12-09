import React, { useState, useEffect, useRef } from "react";
import "./App.css";

function App() {
  const [username, setUsername] = useState("");
  const [connected, setConnected] = useState(false);
  const [board, setBoard] = useState(Array(6).fill(Array(7).fill(0)));
  const [status, setStatus] = useState("Enter username to join");
  const [leaderboard, setLeaderboard] = useState([]);
  const socketRef = useRef(null);

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  const fetchLeaderboard = async () => {
    try {
      const res = await fetch("http://localhost:8080/leaderboard");
      const data = await res.json();
      setLeaderboard(data || []);
    } catch (e) {
      console.error(e);
    }
  };

  const connect = () => {
    if (!username) return;
    setStatus("Searching for opponent...");
    socketRef.current = new WebSocket(
      `ws://localhost:8080/ws?username=${username}`
    );

    socketRef.current.onopen = () => setConnected(true);

    socketRef.current.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === "START") {
        setStatus(`Playing: ${msg.p1} vs ${msg.p2}`);
      } else if (msg.type === "UPDATE") {
        setBoard(msg.board);
        setStatus(
          msg.turn === 0
            ? "Game Over"
            : msg.turn === 1
            ? "Red Turn"
            : "Yellow Turn"
        );
      } else if (msg.type === "OVER") {
        setStatus(`Winner: ${msg.winner}`);
        fetchLeaderboard(); // Refresh stats
        socketRef.current.close();
      }
    };
  };

  const dropDisc = (colIndex) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ col: colIndex }));
    }
  };

  return (
    <div className="App">
      <h1>4 in a Row</h1>

      {!connected ? (
        <div>
          <input
            placeholder="Username"
            onChange={(e) => setUsername(e.target.value)}
          />
          <button onClick={connect}>Find Match</button>
        </div>
      ) : (
        <h3>{status}</h3>
      )}

      <div className="board">
        {board.map((row, rIndex) =>
          row.map((cell, cIndex) => (
            <div
              key={`${rIndex}-${cIndex}`}
              className={`cell ${cell === 1 ? "p1" : cell === 2 ? "p2" : ""}`}
              onClick={() => dropDisc(cIndex)}
            />
          ))
        )}
      </div>

      <div className="leaderboard">
        <h3>Top Players</h3>
        <ul>
          {leaderboard.map((u, i) => (
            <li key={i}>
              {u.username}: {u.wins} wins
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

export default App;
