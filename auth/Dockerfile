FROM golang:1.24-alpine AS builder

WORKDIR /markdown-auth
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download
COPY . .
RUN go build -o auth .

FROM alpine:latest

ARG AUTH_HOST
ARG AUTH_PORT
ENV HOST=${AUTH_HOST}
ENV PORT=${AUTH_PORT}
EXPOSE ${PORT}

WORKDIR /root/
COPY --from=builder /markdown-auth/auth .
CMD ./auth --port=${PORT} --host=${HOST}
