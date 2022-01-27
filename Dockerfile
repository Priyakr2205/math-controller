#syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /math-controller

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY pkg ./pkg/

RUN go build -o /controller

CMD [ "/controller" ]
