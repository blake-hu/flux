FROM golang:1.22.1

WORKDIR /usr/go-backend

COPY ./go-backend .
RUN go mod download

RUN go build -o server

CMD ["./server"]
