package houseband

import (
	"fmt"

	"github.com/kkdai/youtube/v2"
)

type songRequest struct {
	playURL          string
	title            string
	requestChannelID string
}

func newSongRequest(url, channel string) (request songRequest) {
	client := youtube.Client{}
	video, err := client.GetVideo(url)
	audioFormatUrl := video.Formats.FindByItag(251).URL
	if err != nil {
		fmt.Println("Error while getting video: ", err)
	}
	request = songRequest{audioFormatUrl, video.Title, channel}
	return
}
