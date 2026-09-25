# PulseVote Deployment Guide (Render + Vercel)

This guide walks you through deploying **PulseVote** to **Render** (Backend) and **Vercel** (Frontend).

---

## Architecture Overview

- **Backend (Render)**: Go (Gin, WebSocket, Redis Pub/Sub, MongoDB Driver)
- **Frontend (Vercel)**: React + Vite SPA
- **Database (MongoDB Atlas)**: Free Cloud MongoDB
- **Cache & Pub/Sub (Redis)**: Render Redis or Upstash Free Redis

---

## Step 1: Set up Cloud Databases (Free)

### 1. MongoDB Atlas (Database)
1. Go to [MongoDB Atlas](https://www.mongodb.com/cloud/atlas) and sign in/up.
2. Create a **Free Shared Cluster (M0)**.
3. Under **Security > Database Access**, add a user with read/write privileges (note username and password).
4. Under **Security > Network Access**, click **Add IP Address** -> select **Allow Access from Anywhere (`0.0.0.0/0`)** (required for Render to connect).
5. Click **Connect** > **Drivers** (Go) > copy your connection string:
   ```
   mongodb+srv://<username>:<password>@cluster0.xxxxx.mongodb.net/?retryWrites=true&w=majority
   ```

### 2. Redis (Live Vote Counter & Realtime Pub/Sub)
Choose either:
- **Render Redis**: In Render Dashboard, click **New +** > **Redis** (name it `pulsevote-redis`). Copy the connection URL (`redis://...` or `rediss://...`).
- **OR Upstash Redis (Free)**: Go to [upstash.com](https://upstash.com), create a free Redis database, and copy the `rediss://...` connection URL.

---

## Step 2: Push Your Project to GitHub

Open terminal in the project directory:

```bash
git init
git add .
git commit -m "feat: setup pulsevote for Render and Vercel deployment"
git branch -M main
git remote add origin https://github.com/<your-username>/<your-repo-name>.git
git push -u origin main
```

---

## Step 3: Deploy Backend on Render

1. Go to [dashboard.render.com](https://dashboard.render.com) and log in.
2. Click **New +** > **Web Service**.
3. Select **Build and deploy from a Git repository** and connect your GitHub repo.
4. Configure the settings:
   - **Name**: `pulsevote-backend` (or your choice)
   - **Region**: Closest to you (e.g., Singapore, Frankfurt, Oregon)
   - **Root Directory**: `backend` (if repo root is the live-polling-tool folder)
   - **Runtime**: **Docker** (recommended, uses the provided `backend/Dockerfile`)
     *(Alternatively, if choosing Go runtime: Build command: `go build -o server ./cmd/server`, Start command: `./server`)*
   - **Instance Type**: **Free**
5. Add the **Environment Variables**:
   | Key | Value | Description |
   |-----|-------|-------------|
   | `GIN_MODE` | `release` | Production mode |
   | `MONGO_URI` | `mongodb+srv://...` | Your Atlas MongoDB URI |
   | `MONGO_DB` | `pulsevote` | Database name |
   | `REDIS_URL` | `rediss://...` or `redis://...` | Your Redis URI |
   | `JWT_SECRET` | *(Random 32-character secret)* | Secret for auth tokens |
   | `CORS_ORIGIN` | `*` (or your Vercel URL later) | Allowed origin |

6. Click **Create Web Service**.
7. Once deployment finishes, copy your backend URL (e.g. `https://pulsevote-backend.onrender.com`).
8. Verify it works by opening in your browser:
   `https://pulsevote-backend.onrender.com/api/health`
   You should see:
   ```json
   {"service":"PulseVote Live Polling API","status":"ok"}
   ```

---

## Step 4: Deploy Frontend on Vercel

1. Go to [vercel.com](https://vercel.com) and log in.
2. Click **Add New...** > **Project**.
3. Import your GitHub repository.
4. In the project configuration:
   - **Framework Preset**: **Vite**
   - **Root Directory**: Click *Edit* and select `frontend`
   - **Build Command**: `npm run build` (default)
   - **Output Directory**: `dist` (default)
5. Expand **Environment Variables**:
   | Key | Value |
   |-----|-------|
   | `VITE_API_URL` | `https://pulsevote-backend.onrender.com` *(paste your Render backend URL)* |
6. Click **Deploy**.
7. In ~1 minute, Vercel will give you a live production link (e.g., `https://pulsevote-frontend.vercel.app`).

---

## Step 5: (Optional) Update Backend CORS

In Render dashboard:
- Go to `pulsevote-backend` > **Environment**.
- Update `CORS_ORIGIN` to your new Vercel domain: `https://pulsevote-frontend.vercel.app`.
- Save changes (Render will automatically redeploy).

---

## Testing Live Functionality

1. Open your Vercel URL in your browser.
2. Sign up / Log in to create a poll.
3. Open the public poll link in an incognito window or mobile device.
4. Cast a vote — watch the results update in realtime across all devices via WebSockets!
