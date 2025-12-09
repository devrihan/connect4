import React, { useState, useEffect, useRef } from "react";
import "./App.css";

const ClockIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <circle cx="12" cy="12" r="10"></circle>
    <polyline points="12 6 12 12 16 14"></polyline>
  </svg>
);
const ChartIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="20"
    height="20"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <line x1="18" y1="20" x2="18" y2="10"></line>
    <line x1="12" y1="20" x2="12" y2="4"></line>
    <line x1="6" y1="20" x2="6" y2="14"></line>
  </svg>
);
const CloseIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width="24"
    height="24"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <line x1="18" y1="6" x2="6" y2="18"></line>
    <line x1="6" y1="6" x2="18" y2="18"></line>
  </svg>
);

const INITIAL_BOARD = Array(6)
  .fill(0)
  .map(() => Array(7).fill(0));

const AnalyticsModal = ({ onClose }) => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("http://localhost:8081/stats")
      .then((res) => res.json())
      .then((data) => {
        setData(data);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Failed to fetch analytics", err);
        setLoading(false);
      });
  }, []);

  if (!data && !loading) return null;

  return (
    <div className="modal-overlay">
      <div className="modal-content">
        <div className="modal-header">
          <h2>📊 Game Analytics</h2>
          <button className="close-btn" onClick={onClose}>
            <CloseIcon />
          </button>
        </div>

        {loading ? (
          <p>Loading stats...</p>
        ) : (
          <div className="stats-grid">
            <div className="stat-card">
              <h4>Total Games</h4>
              <p className="stat-value">{data.totalGames}</p>
            </div>
            <div className="stat-card">
              <h4>Avg Duration</h4>
              <p className="stat-value">{data.avgDuration?.toFixed(1)}s</p>
            </div>

            <div className="stat-full">
              <h4>🏆 Top Winners</h4>
              <ul className="stat-list">
                {Object.entries(data.winsPerUser || {}).map(([user, wins]) => (
                  <li key={user}>
                    <span>{user}</span>
                    <span className="bold">{wins} wins</span>
                  </li>
                ))}
                {Object.keys(data.winsPerUser || {}).length === 0 && (
                  <li className="empty">No data yet</li>
                )}
              </ul>
            </div>

            <div className="stat-full">
              <h4>⏰ Activity (Games per Hour)</h4>
              <div className="activity-bar-container">
                {Object.entries(data.gamesPerHour || {}).map(
                  ([hour, count]) => (
                    <div key={hour} className="activity-item">
                      <div
                        className="activity-bar"
                        style={{ height: `${Math.min(count * 10, 50)}px` }}
                      ></div>
                      <span className="activity-label">{hour}:00</span>
                    </div>
                  )
                )}
                {Object.keys(data.gamesPerHour || {}).length === 0 && (
                  <p className="empty">No activity yet</p>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

function App() {
  const [username, setUsername] = useState("");
  const [connected, setConnected] = useState(false);
  const [board, setBoard] = useState(INITIAL_BOARD);
  const [currentTurn, setCurrentTurn] = useState(0);
  const [p1Name, setP1Name] = useState(null);
  const [p2Name, setP2Name] = useState(null);
  const [localPlayerID, setLocalPlayerID] = useState(0);
  const [winner, setWinner] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);
  const [showAnalytics, setShowAnalytics] = useState(false);
  const socketRef = useRef(null);
  const previousBoard = useRef(INITIAL_BOARD);
  const [lastMove, setLastMove] = useState(null);

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  useEffect(() => {
    if (!connected || currentTurn === 0) return;
    if (!Array.isArray(board) || !Array.isArray(previousBoard.current)) return;

    const flatBoard = board.flat();
    const flatPrevBoard = previousBoard.current.flat();

    let changedCell = -1;
    for (let i = 0; i < flatBoard.length; i++) {
      if (flatBoard[i] !== flatPrevBoard[i] && flatBoard[i] !== 0) {
        changedCell = i;
        break;
      }
    }

    if (changedCell !== -1) {
      const row = Math.floor(changedCell / 7);
      const col = changedCell % 7;
      const player = board[row][col];
      setLastMove({ row, col, player });
      const discColor = player === 1 ? "#c0392b" : "#d4ac0d";
      document.documentElement.style.setProperty("--disc-color", discColor);
      setTimeout(() => setLastMove(null), 500);
    }
    previousBoard.current = board.map((row) => [...row]);
  }, [board, connected, currentTurn]);

  const fetchLeaderboard = async () => {
    try {
      const res = await fetch("http://localhost:8080/leaderboard");
      const data = await res.json();
      setLeaderboard(Array.isArray(data) ? data : []);
    } catch (e) {
      console.error(e);
    }
  };

  const connect = () => {
    if (!username) return;
    setConnected(false);
    setBoard(INITIAL_BOARD);
    setCurrentTurn(0);
    setWinner(null);
    setLocalPlayerID(0);
    setP1Name(null);
    setP2Name(null);
    setLastMove(null);

    socketRef.current = new WebSocket(
      `ws://localhost:8080/ws?username=${username}`
    );
    socketRef.current.onopen = () => setConnected(true);
    socketRef.current.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === "START") {
        setP1Name(msg.p1);
        setP2Name(msg.p2);
        const myID = msg.p1 === username ? 1 : 2;
        setLocalPlayerID(myID);
        setCurrentTurn(1);
        previousBoard.current = INITIAL_BOARD;
      } else if (msg.type === "UPDATE") {
        if (Array.isArray(msg.board) && msg.board.length === 6)
          setBoard(msg.board);
        setCurrentTurn(msg.turn);
      } else if (msg.type === "OVER") {
        setWinner(msg.winner);
        setCurrentTurn(0);
        if (Array.isArray(msg.board) && msg.board.length === 6)
          setBoard(msg.board);
        fetchLeaderboard();
      }
    };
    socketRef.current.onclose = () => {
      setConnected(false);
      if (!winner) setBoard(INITIAL_BOARD);
    };
  };

  const dropDisc = (colIndex) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ col: colIndex }));
    }
  };

  return (
    <div className="App">
      <h1 className="title">4 in a Row</h1>

      {showAnalytics && (
        <AnalyticsModal onClose={() => setShowAnalytics(false)} />
      )}

      {!connected || !p1Name ? (
        <div className="connection-form">
          <input
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
          <button onClick={connect}>Find Match</button>
        </div>
      ) : null}

      {connected && p1Name && (
        <div className="status-bar">
          <div
            className={`player-info ${
              currentTurn === 1 ? "active" : "inactive"
            }`}
          >
            <div className="player-dot p1"></div>
            <div className="player-details">
              <p className="player-name">{p1Name}</p>
              <p className="player-color">(Red)</p>
            </div>
          </div>
          <div className="turn-indicator">
            {currentTurn === localPlayerID ? (
              <span className="your-turn-text">Your Turn!</span>
            ) : (
              <span className="waiting-text">
                <ClockIcon /> Waiting...
              </span>
            )}
          </div>
          <div
            className={`player-info right ${
              currentTurn === 2 ? "active" : "inactive"
            }`}
          >
            <div className="player-details">
              <p className="player-name">{p2Name}</p>
              <p className="player-color">(Yellow)</p>
            </div>
            <div className="player-dot p2"></div>
          </div>
        </div>
      )}

      {winner && (
        <h2 className="status-message winner-text">Winner: {winner}</h2>
      )}

      {connected && p1Name && (
        <div className="board">
          {board?.map((row, rIndex) =>
            row?.map((cell, cIndex) => {
              const isLastMove =
                lastMove && lastMove.row === rIndex && lastMove.col === cIndex;
              return (
                <div
                  key={`${rIndex}-${cIndex}`}
                  className={`cell ${
                    cell === 1 ? "p1" : cell === 2 ? "p2" : ""
                  } ${isLastMove ? "falling-disc" : ""}`}
                  style={{
                    "--target-row": rIndex,
                    cursor:
                      connected &&
                      !winner &&
                      currentTurn === localPlayerID &&
                      board[0]?.[cIndex] === 0
                        ? "pointer"
                        : "default",
                  }}
                  onClick={() => {
                    if (connected && !winner && currentTurn === localPlayerID)
                      dropDisc(cIndex);
                  }}
                />
              );
            })
          )}
        </div>
      )}

      <div className="leaderboard-container">
        <h3 className="leaderboard-title">🏆 Top Players</h3>
        <ul className="leaderboard-list">
          {leaderboard.map((u, i) => (
            <li
              key={i}
              className="leaderboard-item"
              style={{
                backgroundColor: u.username === username ? "#4a4f66" : "",
              }}
            >
              <span className="rank">{i + 1}.</span>
              <span className="username">{u.username}</span>
              <span className="wins">{u.wins} wins</span>
            </li>
          ))}
        </ul>
      </div>

      <button
        className="analytics-btn"
        onClick={() => setShowAnalytics(true)}
        style={{ marginTop: "25px" }}
      >
        <ChartIcon /> Analytics
      </button>
    </div>
  );
}

export default App;
