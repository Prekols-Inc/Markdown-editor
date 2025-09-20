FROM golang:1.24.0 AS builder
WORKDIR /markdown-auth
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth .

FROM alpine:latest
COPY .env . 
WORKDIR /root/
COPY --from=builder /markdown-auth/auth .
CMD ["./auth"]
