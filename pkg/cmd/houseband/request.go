package houseband

import (
	"fmt"

	"github.com/kkdai/youtube/v2"
)

type ytRequest struct {
	*youtube.Video
	nowPlaying func()
}

func newYtRequest(url string, callback func(string)) (request ytRequest) {
	client := youtube.Client{}
	video, err := client.GetVideo(url)
	//audioFormatUrl := video.Formats.FindByItag(251).URL
	if err != nil {
		fmt.Println("Error while getting video: ", err)
	}
	nowPlaying := func() { callback("**Now Playing:** `" + video.Title + "`") }
	request = ytRequest{video, nowPlaying}
	return
}
