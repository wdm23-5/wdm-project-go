FROM golang:1.19

WORKDIR /home/payment-service

COPY common common
COPY main/payment.go main/payment.go
COPY payment payment
COPY go.mod go.sum ./

RUN go build -ldflags="-s -w" -v -o payment-gin main/payment.go

EXPOSE 5000
