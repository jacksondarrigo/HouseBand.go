package houseband

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

type request struct {
	*youtube.Video
	streamURL  string
	nowPlaying func()
}

func newRequest(video *youtube.Video, streamURL, channelID string, callback func(string, string) (*discordgo.Message, error)) request {
	nowPlaying := func() { callback(channelID, "**Now Playing:** `"+video.Title+"`") }
	return request{video, streamURL, nowPlaying}
}
