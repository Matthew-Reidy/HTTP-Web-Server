FROM golang:alpine

COPY . /app

WORKDIR /app

CMD go build main.go; ./main