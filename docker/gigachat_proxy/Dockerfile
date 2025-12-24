FROM golang:1.24-alpine AS builder

WORKDIR /markdown-gigachat-proxy
COPY gigachat_proxy/go.mod gigachat_proxy/go.sum .
RUN go mod tidy
RUN go mod download
COPY ./gigachat_proxy .
RUN go build -o gigachat_proxy .

FROM alpine:latest

EXPOSE ${GIGACHAT_PROXY_PORT}

WORKDIR /root/
COPY --from=builder /markdown-gigachat-proxy/gigachat_proxy .
CMD ./gigachat_proxy --port=${GIGACHAT_PROXY_PORT} --host=${GIGACHAT_PROXY_HOST}
