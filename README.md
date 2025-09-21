# Markdown Editor

## Build

### Frontend

> Prerequisites: **Node.js ≥ 18 LTS** and **npm** (or `pnpm` / `yarn`).

```bash
# 1. Go to the frontend folder
cd frontend

# 2. Install dependencies
npm install

# 3. Add VITE_AUTH_API_BASE_URL env variable in frontend/.env file
# Example: VITE_AUTH_API_BASE_URL=http://localhost:8080

# 4. Start the dev server
npm run dev
```

Admin credentials:
- Username: `admin`
- Password: `password`
---

###  Backend

```bash
cd backend
go run main.go
```

### Auth service

> Prerequisites: **Go ≥ 1.23**.

```bash
# 1. Go to the frontend folder
cd auth

# 2. Get dependencies
go mod tidy

# 3 Start the auth service
go run .
```

### Docker

```bash
# 1. Авторизация в Docker Hub
docker login

# 2. Создание файла .env в корневой директории проекта
cat > .env << EOF
BACKEND_HOST=localhost
BACKEND_PORT=1234
AUTH_HOST=localhost
AUTH_PORT=8080
FRONTEND_HOST=localhost
FRONTEND_PORT=5173
DOCKERHUB_USERNAME=[your-dockerhub-username]
EOF

# 3. Сборка образов
docker compose build

# 4. Пуш образов в Docker Hub
docker compose push

# 5. Запуск приложения
docker-compose up -d

---
## Tests
### Test backend
```bash
cd backend
go test -v
```

### Test auth service
```bash
cd auth

go test
```
