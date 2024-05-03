FROM golang:1.22.1

COPY ./go-backend .
RUN go mod download

RUN go build -o server

CMD ["./server"]
