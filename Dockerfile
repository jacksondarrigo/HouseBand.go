FROM golang:1.19 AS builder

RUN mkdir /go/src/houseband
WORKDIR /go/src/houseband

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/houseband ./cmd/houseband

FROM python:3.10-slim
RUN apt-get update && apt-get install -y ffmpeg
ADD https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp
COPY --from=builder /go/bin/houseband /go/bin/houseband
ENTRYPOINT ["/go/bin/houseband"]
