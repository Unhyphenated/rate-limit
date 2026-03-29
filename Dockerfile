FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.sum go.mod ./

RUN go mod download

COPY . .

RUN go build -o server ./cmd/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]