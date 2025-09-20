FROM golang:1.24.0 AS builder

WORKDIR /markdown-backend
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o backend .

FROM alpine:latest
COPY .env . 
WORKDIR /root/
COPY --from=builder /markdown-backend/backend .
CMD ["./backend"]
