package request

import (
	"fmt"

	"github.com/kkdai/youtube/v2"
)

type Request struct {
	*youtube.Video
	StreamURL  string
	NowPlaying func()
}

var youtubeClient *youtube.Client = &youtube.Client{}

func New(url string, callback func(title string)) (req *Request, err error) {
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
	return &Request{video, stream, func() { callback(video.Title) }}, nil
}
