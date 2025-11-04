FROM golang:1.24-alpine AS builder

WORKDIR /markdown-backend
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download
COPY . .
RUN go build -o backend .

FROM alpine:latest

EXPOSE ${BACKEND_PORT}

WORKDIR /root/
COPY --from=builder /markdown-backend/backend .
COPY ./docs ./docs
COPY ./tls ./tls
CMD ./backend --port=${BACKEND_PORT} --host=${BACKEND_HOST}
