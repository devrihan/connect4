# Connect4

A full-stack implementation of the classic **Connect Four** game, built with a modular architecture featuring a client, server, and analytics layer.  
This project is containerized using Docker Compose for easy setup and deployment.



##  Project Structure

```
connect4/
│
├── client/             # Frontend application
├── server/             # Backend (API and game logic)
├── analytics/          # Data collection and analysis
├── docker-compose.yml  # Docker setup for multi-service deployment
└── .gitignore
```



##  Features

- Multiplayer **Connect 4** gameplay
- Web-based interface (client)
- Real-time game state management (server)
- Analytics tracking for game sessions
- Docker-based setup for seamless development and deployment


##  Setup & Installation

### 1. Clone the Repository
```bash
git clone https://github.com/devrihan/connect4.git
cd connect4
```

### 2. Run Using Docker
```bash
docker-compose up --build
```

This command will:
- Build and run the client (frontend)
- Launch the server (backend API)
- Start the analytics service

Once built, the application should be available at:
```
http://localhost:3000
```


##  Tech Stack

| Layer | Technology |
|-------|-------------|
| Client | React / Vite (or similar frontend framework) |
| Server | Node.js / Express |
| Analytics | Python or Node-based data processing |
| Deployment | Docker & Docker Compose |


##  Development

To run components individually:

### Client
```bash
cd client
npm install
npm start
```

### Server
```bash
cd server
npm install
npm run dev
```

### Analytics
```bash
cd analytics
python main.py
```


##  Analytics

The analytics service collects and processes gameplay data for:
- Player performance
- Move efficiency
- Game duration statistics


##  Contact

**Author:** [@devrihan](https://github.com/devrihan)  
**Repository:** [connect4](https://github.com/devrihan/connect4)
