# Markdown Editor

## Build

The most easiest way to set up the project is Docker Compose:

1. Create `.env` file like `.env.example`

2. Run project:
 - To run all containers in a background (detach process)
```bash
docker-compose up -d
```
 - To attach all processes to the terminal session
```bash
docker compose up
```

### Build from source

#### Frontend

> Prerequisites: **Node.js ≥ 18 LTS** and **npm** (or `pnpm` / `yarn`).

```bash
# 1. Go to the frontend folder
cd frontend

# 2. Install dependencies
npm install

# 3. Add env variables in frontend/.env file:
# Example: 
VITE_AUTH_API_BASE_URL=http://localhost:8080
VITE_BACKEND_API_BASE_URL=http://localhost:1234

# 4. Start the dev server
npm run dev
```

Admin credentials:
- Username: `admin`
- Password: `password`
---

####  Backend

```bash
$ cd backend
$ go mod tidy
$ go run . --host 
```

#### Auth service

> Prerequisites: **Go ≥ 1.23**.

```bash
$ cd auth
$ go mod tidy
$ go run .
```

---
## Tests
### Backend
```bash
cd backend
go test -v
```

### Auth service
```bash
cd auth
go test -v
```
