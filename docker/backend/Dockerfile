FROM golang:1.25-alpine AS builder

WORKDIR /markdown-backend/backend
COPY backend/go.mod backend/go.sum .
RUN go mod tidy
RUN go mod download
COPY backend .
COPY lib ../lib
RUN go build -o backend .

FROM alpine:latest

EXPOSE ${BACKEND_PORT}

WORKDIR /root/
COPY --from=builder /markdown-backend/backend .
COPY backend/docs docs
CMD ./backend --port=${BACKEND_PORT} --host=${BACKEND_HOST}
