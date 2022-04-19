package houseband

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
)

type request struct {
	*youtube.Video
	nowPlaying func()
}

func newRequest(video *youtube.Video, channelID string, callback func(string, string) (*discordgo.Message, error)) request {
	nowPlaying := func() { callback(channelID, "**Now Playing:** `"+video.Title+"`") }
	return request{video, nowPlaying}
}
