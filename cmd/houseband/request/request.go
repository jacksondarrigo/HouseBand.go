package request

import (
	"os/exec"
	"strings"
)

type Request struct {
	ReqURL string
	Title  string
	// StreamURL string
	nowPlaying chan bool
}

func New(url string, callback chan bool) (req *Request, err error) {
	title, err := exec.Command("youtube-dl", "-e", url).Output()
	if err != nil {
		return nil, err
	}
	return &Request{url, strings.TrimSuffix(string(title), "\n"), callback}, nil
}

func (r Request) GetStream() (string, error) {
	streamUrl, err := exec.Command("youtube-dl", "-f", "bestaudio", "-g", r.ReqURL).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(streamUrl), "\n"), nil
}

func (r Request) NowPlaying() {
	r.nowPlaying <- true
}

func (r Request) Cancel() {
	r.nowPlaying <- false
}
