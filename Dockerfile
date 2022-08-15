FROM golang:1.19 AS builder

RUN mkdir /go/src/houseband
WORKDIR /go/src/houseband

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/houseband ./cmd/houseband

FROM ubuntu:focal
RUN apt-get update && apt-get install -y ffmpeg ca-certificates
COPY --from=builder /go/bin/houseband /go/bin/houseband
ENTRYPOINT ["/go/bin/houseband"]
