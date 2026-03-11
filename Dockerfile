FROM golang:1.25.6-alpine AS builder
WORKDIR /app

# Which service folder to build (e.g. order_api, order_processor, payment_service)
ARG SERVICE_DIR
COPY go.work go.work.sum ./
COPY shared ./shared
COPY services/order_api ./services/order_api
COPY services/order_processor ./services/order_processor
COPY services/payment_service ./services/payment_service
RUN go build -trimpath -ldflags="-s -w" -o /app/service "./${SERVICE_DIR}"

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/service .
CMD ["./service"]