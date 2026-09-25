# PulseVote – Live Polling Tool

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.10.0-008ECF?style=flat&logo=go)](https://gin-gonic.com/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://react.dev)
[![Vite](https://img.shields.io/badge/Vite-5-646CFF?style=flat&logo=vite)](https://vitejs.dev)
[![Redis](https://img.shields.io/badge/Redis-Live%20Counts%20%26%20Pub%2FSub-DC382D?style=flat&logo=redis)](https://redis.io)
[![MongoDB](https://img.shields.io/badge/MongoDB-Permanent%20Storage-47A248?style=flat&logo=mongodb)](https://mongodb.com)
[![WebSockets](https://img.shields.io/badge/WebSockets-Gorilla-black?style=flat)](https://github.com/gorilla/websocket)

> **PulseVote** is a production-grade, real-time live polling platform. It empowers presenters, educators, and event hosts to create live polls, share an instant link, and watch thousands of audience votes update dynamically across all connected screens **without ever refreshing the page**.

---

## Architecture Role Breakdown

PulseVote implements a clear separation of concerns across every layer of the stack:

| Technology | Responsibility | Details |
| :--- | :--- | :--- |
| **React + Vite** | **Frontend Client** | Responsive UI, state management, live SVG indicators, animated progress bars, WebSocket listeners. |
| **Go + Gin** | **Backend API & WebSockets** | High-concurrency RESTful API, JWT auth, request validation, Gorilla WebSocket hub connection manager. |
| **MongoDB** | **Permanent Data Store** | Persistent ACID records for users, polls, options, and permanent vote audit history in `pulsevote` database. |
| **Redis (Hash)** | **Live In-Memory Counters** | Sub-millisecond atomic counter incrementation (`HINCRBY poll:<shareCode>:results <optionId> 1`). |
| **Redis (Pub/Sub)** | **Message Broker** | Cross-instance event bus publishing vote delta messages to channel `poll:<shareCode>`. |
| **Gorilla WebSocket** | **Real-Time Client Broadcast** | Manages concurrent client sockets and pushes live count updates directly to active browser tabs. |

---

## Real-Time WebSocket Flow

When an audience member casts a vote, the following real-time sequence executes:

```
[Audience Window 1: Clicks Vote Option]
             │
             ▼
[HTTP POST /api/polls/:shareCode/vote]
             │
             ├──► 1. Save Vote permanently to MongoDB (Audit & Durability)
             │
             ├──► 2. Execute HINCRBY in Redis Hash (Atomic live counter)
             │
             └──► 3. Publish payload to Redis Pub/Sub channel "poll:<shareCode>"
                                    │
                                    ▼
                 [Redis Subscriber in Go Backend]
                                    │
                                    ▼
           [Gorilla WebSocket Hub: clients[shareCode]]
             │                                     │
             ▼                                     ▼
[Window 1: Instant Feedback]           [Window 2: Screen updates in real-time]
  "Vote submitted successfully"           Live counts, % and progress bars animate!
```

---

## Features

- **Authentication & Security**:
  - Secure signup and login with email validation.
  - Salted password hashing with `bcrypt`.
  - Stateless JWT token authorization (`Authorization: Bearer <token>`).
  - Strict ownership checks: only poll creators can close, edit, or delete their polls.
- **Poll Management**:
  - Create polls with 2 to 10 distinct options.
  - Generates a unique, collision-resistant 6-character public share code.
  - Dashboard displaying all polls, active states, option counts, and creation dates.
  - Ability to close polls (disables further voting while keeping live results accessible).
  - Ability to permanently delete polls along with their MongoDB records and Redis cache.
- **Audience Voting (No Login Required)**:
  - Clean public voting page at `/poll/:shareCode`.
  - Single-click interactive voting buttons.
  - Client vote tracking prevents multiple accidental votes.
  - Live animated percentage progress bars with smooth CSS transitions.
- **Real-Time Synchronization**:
  - Gorilla WebSocket hub maintains active connection pools per poll.
  - Redis Pub/Sub guarantees instantaneous distribution across multiple client windows.
  - Live status indicator (🟢 *Live updating without reload*).

---

## Folder Structure

```
live-polling-tool/
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Navbar.jsx              # Navigation header with auth controls
│   │   │   ├── Footer.jsx              # Application footer
│   │   │   ├── ProgressBar.jsx         # Live animated percentage progress bar
│   │   │   ├── StatusBadge.jsx         # Active (pulsing) / Closed badge
│   │   │   ├── ShareModal.jsx          # Copy link & new tab popup
│   │   │   └── DeleteConfirmModal.jsx  # Deletion confirmation dialog
│   │   ├── pages/
│   │   │   ├── Home.jsx                # Landing page with interactive live simulator
│   │   │   ├── Signup.jsx              # User registration
│   │   │   ├── Login.jsx               # User authentication
│   │   │   ├── Dashboard.jsx           # Poll creator dashboard & management
│   │   │   ├── CreatePoll.jsx          # Dynamic 2-10 option poll builder
│   │   │   └── PollPage.jsx            # Public voting & real-time results page
│   │   ├── services/
│   │   │   ├── api.js                  # Centralized HTTP API client
│   │   │   └── websocket.js            # Gorilla WebSocket connection manager
│   │   ├── hooks/
│   │   │   └── usePoll.js              # Real-time state & WebSocket event hook
│   │   ├── context/
│   │   │   └── AuthContext.jsx         # Session and JWT persistence
│   │   ├── App.jsx                     # Route definitions & guards
│   │   ├── main.jsx                    # React 18 DOM root mounting
│   │   └── index.css                   # Custom design system & theme tokens
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   ├── .env
│   └── .env.example
│
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                 # App entrypoint, graceful shutdown, server wiring
│   ├── config/
│   │   ├── database.go                 # MongoDB client & collection index init
│   │   └── redis.go                    # Redis client initialization & ping check
│   ├── controllers/
│   │   ├── auth.go                     # Signup & Login HTTP handlers
│   │   ├── poll.go                     # Poll CRUD & results HTTP handlers
│   │   └── vote.go                     # Vote submission HTTP handler
│   ├── middleware/
│   │   └── auth.go                     # Bearer JWT verification middleware
│   ├── models/
│   │   ├── user.go                     # User struct & Auth DTOs
│   │   ├── poll.go                     # Poll, Option, and Result DTOs
│   │   └── vote.go                     # Vote & Pub/Sub message definitions
│   ├── repository/
│   │   ├── user_repository.go          # MongoDB user CRUD operations
│   │   └── poll_repository.go          # MongoDB poll & vote aggregation operations
│   ├── services/
│   │   ├── auth.go                     # Password hashing & JWT generation
│   │   ├── poll.go                     # Poll creation, option validation & vote logic
│   │   └── realtime.go                 # Redis hash increment & Pub/Sub subscriber
│   ├── websocket/
│   │   └── handler.go                  # Gorilla WebSocket Hub, read/write pumps
│   ├── utils/
│   │   ├── jwt.go                      # HMAC SHA256 JWT claims generator & validator
│   │   └── jwt_test.go                 # JWT unit tests
│   ├── .env
│   ├── .env.example
│   ├── go.mod
│   └── go.sum
│
├── docker-compose.yml                  # One-command MongoDB & Redis startup
├── test_plan.ps1                       # Automated 16-step test suite (PowerShell)
├── test_plan.sh                        # Automated 16-step test suite (Bash)
├── README.md                           # Documentation
└── .gitignore
```

---

## Environment Variables

### Backend (`backend/.env`)

```env
PORT=8080
GIN_MODE=debug

# MongoDB Configuration
# Local MongoDB:
MONGO_URI=mongodb://localhost:27017
# Or MongoDB Atlas:
# MONGO_URI=mongodb+srv://<username>:<password>@cluster0.mongodb.net/?retryWrites=true&w=majority
MONGO_DB=pulsevote

# Redis Configuration
# Local Redis:
REDIS_URL=redis://localhost:6379
# Or Redis Cloud / Upstash:
# REDIS_URL=rediss://default:<password>@<host>:<port>

# JWT Authentication Secret
JWT_SECRET=pulsevote_super_secret_jwt_key_2026_change_in_production

# Frontend Origin for CORS
CORS_ORIGIN=http://localhost:5173
```

### Frontend (`frontend/.env`)

```env
# URL where PulseVote Go backend is running
VITE_API_URL=http://localhost:8080
```

---

## Prerequisites & Installation

### 1. Start MongoDB and Redis

You can start both instantly using the included Docker Compose configuration:

```bash
docker compose up -d
```

*Or use free cloud instances:*
- **MongoDB Atlas**: Free M0 Sandbox cluster at [mongodb.com/atlas](https://www.mongodb.com/atlas)
- **Redis Cloud / Upstash**: Free cloud Redis at [upstash.com](https://upstash.com) or [redis.io/cloud](https://redis.io/cloud)

### 2. Setup & Run Backend (Go)

```bash
cd backend

# Download dependencies
go mod tidy

# Run server
go run cmd/server/main.go
```

The backend server starts listening on `http://localhost:8080`.

### 3. Setup & Run Frontend (React + Vite)

In a separate terminal:

```bash
cd frontend

# Install npm packages
npm install

# Start development server
npm run dev
```

The frontend application opens on `http://localhost:5173`.

---

## API Endpoints Reference

### Health Check
- `GET /api/health` – Returns `{ "status": "ok", "service": "PulseVote Live Polling API" }`.

### Authentication
- `POST /api/auth/signup` – Register a new account (`name`, `email`, `password`, `confirmPassword`).
- `POST /api/auth/login` – Login with credentials (`email`, `password`), returns JWT and user profile.

### Polls (Protected – Requires `Authorization: Bearer <token>`)
- `POST /api/polls` – Create a new poll (`question`, `options` array [2–10 items]).
- `GET /api/polls` – Retrieve all polls created by the logged-in user.
- `PUT /api/polls/:id` – Update poll question (owner only).
- `DELETE /api/polls/:id` – Delete poll, votes, and Redis cache (owner only).
- `POST /api/polls/:id/close` – Mark poll as closed/inactive (owner only).

### Public Polls & Voting (No Auth Required)
- `GET /api/polls/:shareCode` – Retrieve public poll information and initial option counts.
- `POST /api/polls/:shareCode/vote` – Submit a vote (`{ "optionId": "opt_1" }`).
- `GET /api/polls/:shareCode/results` – Retrieve current live vote counts and calculated percentages.

### WebSocket
- `GET /ws/polls/:shareCode` – Upgrade connection to real-time Gorilla WebSocket stream.

---

## Testing

PulseVote includes both Go unit tests and automated 16-step end-to-end API test suites.

### Running Go Unit Tests
```bash
cd backend
go test -v ./...
```

### Running Automated 16-Step Verification Test
Ensure your backend is running, then execute:

**Windows PowerShell:**
```powershell
.\test_plan.ps1
```

**Linux / macOS:**
```bash
chmod +x test_plan.sh
./test_plan.sh
```

The test script automatically validates:
1. Signup with valid fields
2. Rejection of duplicate email with HTTP 409
3. Login with valid credentials
4. Rejection of invalid credentials with HTTP 401
5. Poll creation with valid 2–10 options
6. Rejection of invalid poll (<2 options) with HTTP 400
7. Public poll retrieval without auth
8. Successful vote submission
9. Rejection of invalid option ID
10. Multiple concurrent votes
11. Verification that Redis live option counts updated
12. Gorilla WebSocket endpoint availability
13. Poll owner closing poll
14. Rejection of votes on closed poll with HTTP 403
15. Viewing final results on closed poll
16. Poll deletion and cache purge

---

## Multi-Window Live Demonstration Walkthrough

To verify and demonstrate real-time WebSocket synchronization:

1. Open your browser and navigate to `http://localhost:5173`.
2. Sign up and log in.
3. Click **Create Poll** and enter:
   - Question: *"Which technology do you prefer for real-time applications?"*
   - Option 1: *Go + Gorilla WebSocket*
   - Option 2: *Node.js + Socket.IO*
   - Option 3: *Elixir + Phoenix Channels*
4. Click **Create Poll** and copy the generated public link (e.g., `http://localhost:5173/poll/a7x9b2`).
5. Open two separate browser windows side by side:
   - **Window 1 (Audience A)**: Paste `http://localhost:5173/poll/a7x9b2`.
   - **Window 2 (Audience B)**: Paste `http://localhost:5173/poll/a7x9b2`.
6. Notice both windows display the green `🟢 Live & Realtime` indicator.
7. Click **Go + Gorilla WebSocket** in **Window 1**.
8. **Watch Window 2 instantly update the vote count, percentage, and animated progress bar without refreshing!**
9. Go back to your Dashboard tab, click **Close** on the poll.
10. Notice Window 1 and Window 2 instantly display *"This poll is closed. Voting is disabled."* in real-time.

---

## Deployment Guide

### Frontend Deployment (Vercel / Netlify / Cloudflare Pages)
1. Push your repository to GitHub.
2. Connect your repo in Vercel or Netlify.
3. Set the Root Directory to `frontend`.
4. Configure Build Command: `npm run build`.
5. Output Directory: `dist`.
6. Add Environment Variable:
   - `VITE_API_URL` = `https://your-backend-domain.com`

### Backend Deployment (Render / Railway / Fly.io / DigitalOcean)
1. Set up a Go service pointing to `backend/`.
2. Build Command: `go build -o server cmd/server/main.go`.
3. Start Command: `./server`.
4. Add Environment Variables:
   - `MONGO_URI` = MongoDB Atlas connection string (`mongodb+srv://...`)
   - `REDIS_URL` = Redis Cloud / Upstash URL (`rediss://...`)
   - `JWT_SECRET` = A strong random secret key
   - `CORS_ORIGIN` = `https://your-frontend-domain.vercel.app`
   - `PORT` = `8080` (or default port provided by host)

---

## Screenshots

*(Mockups and UI snapshots of PulseVote)*

| Landing Page | Live Polling Page |
| :---: | :---: |
| Hero banner with simulated live preview card | Instant one-click voting & real-time progress bars |

| Creator Dashboard | Create Poll Builder |
| :---: | :---: |
| Full poll management, status toggles & links | Dynamic 2–10 option validation & instant share link |

---

## Demo Video & Presentation Guide

For demonstrating PulseVote in an interview or project review:
1. **Show the live split-screen demo first**: Recruiters and engineers love seeing real-time WebSockets work without reloading.
2. **Explain the two-tier storage decision**:
   - Why not just write to MongoDB? MongoDB handles permanent ACID durability, while Redis delivers sub-millisecond atomic counter increments.
   - Why Redis Pub/Sub? In a production multi-replica backend (e.g., 3 Go server instances behind a load balancer), Pub/Sub ensures all instances receive vote events and broadcast them to their respective connected WebSockets.
3. **Showcase Gorilla WebSocket connection resilience**: Clean disconnect handling, ping/pong heartbeats, and client concurrency protection.
