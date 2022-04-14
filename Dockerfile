FROM golang:1.17

RUN apt update
RUN apt install -y ffmpeg

RUN go get golang.org/x/net/html
RUN go get github.com/bwmarrin/discordgo@master

WORKDIR /go/src/houseband
COPY . .

RUN go install ./cmd/houseband/

ENTRYPOINT ["houseband"]
