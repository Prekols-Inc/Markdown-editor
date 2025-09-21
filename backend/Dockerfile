FROM golang:1.24-alpine AS builder

WORKDIR /markdown-backend
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download
COPY . .
RUN go build -o backend .

FROM alpine:latest

ARG BACKEND_HOST
ARG BACKEND_PORT
EXPOSE ${BACKEND_PORT}

WORKDIR /root/
COPY --from=builder /markdown-backend/backend .
CMD ./backend --port=${BACKEND_PORT} --host=${BACKEND_HOST}
