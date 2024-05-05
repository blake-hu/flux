FROM golang:1.22.1

WORKDIR /deploy/go-backend

COPY ./go-backend .
RUN go mod download

RUN go build -o server

CMD ["./server"]
