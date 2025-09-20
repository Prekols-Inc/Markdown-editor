FROM golang:1.24-alpine AS builder

WORKDIR /markdown-backend
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download
COPY . .
RUN go build -o backend .

FROM alpine:latest

ARG BACKEND_PORT
ENV PORT=${BACKEND_PORT}
EXPOSE ${PORT}

WORKDIR /root/
COPY --from=builder /markdown-backend/backend .
CMD ./backend --port=${PORT} --host=${PORT}
