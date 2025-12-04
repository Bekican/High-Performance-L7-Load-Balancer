
FROM golang:1.22-alpine AS builder


WORKDIR /app


COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o loadbalancer cmd/lb/main.go


FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/loadbalancer .

COPY config.yaml .
COPY dashboard.html .

EXPOSE 8080 9091

CMD ["./loadbalancer"]