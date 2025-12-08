FROM golang:1.25-alpine AS builder

WORKDIR /markdown-auth/auth
COPY auth/go.mod auth/go.sum .
RUN go mod tidy
RUN go mod download
COPY auth .
COPY lib ../lib
RUN go build -o auth .

FROM alpine:latest

EXPOSE ${AUTH_PORT}

WORKDIR /root/
COPY --from=builder /markdown-auth/auth .
COPY auth/docs docs
COPY auth/tls tls
CMD ./auth --port=${AUTH_PORT} --host=${AUTH_HOST}
