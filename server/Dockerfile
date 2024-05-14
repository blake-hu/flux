FROM golang:1.22.1

WORKDIR /deploy/server

COPY ./server .
RUN go mod download

RUN go build -o server

CMD ["./server"]
