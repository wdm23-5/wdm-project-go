FROM golang:1.19

WORKDIR /home/stock-service

COPY common common
COPY main/stock.go main/stock.go
COPY stock stock
COPY go.mod go.sum ./

RUN go build -ldflags="-s -w" -v -o stock-gin main/stock.go

EXPOSE 5000
