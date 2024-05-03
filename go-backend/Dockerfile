FROM golang:1.22.1

WORKDIR /app

copy . .
RUN go mod download

RUN go build -o server ./main.go

cmd ["./server"]


