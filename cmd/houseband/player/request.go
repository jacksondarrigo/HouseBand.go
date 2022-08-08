package player

import (
	"fmt"

	"github.com/kkdai/youtube/v2"
)

type request struct {
	*youtube.Video
	streamURL  string
	nowPlaying func()
}

var youtubeClient *youtube.Client = &youtube.Client{}

func NewRequest(url string, callback func(title string)) (req *request, err error) {
	video, err := youtubeClient.GetVideo(url)
	if err != nil {
		fmt.Println("Error while getting video: ", err)
		return nil, err
	}
	stream, err := youtubeClient.GetStreamURL(video, video.Formats.FindByItag(251))
	if err != nil {
		fmt.Println("Error while getting stream URL: ", err)
		return nil, err
	}
	return &request{video, stream, func() { callback(video.Title) }}, nil
}
