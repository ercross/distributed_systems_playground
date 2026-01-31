FROM golang:1.25.6-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o service main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/service .

CMD ["./service"]