package request

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

func isValidURL(reqUrl string) bool {
	// Check it's an Absolute URL or absolute path
	uri, err := url.Parse(reqUrl)
	if err != nil {
		return false
	}
	// Check it's a valid domain name
	_, err = net.LookupHost(uri.Host)
	return err == nil
}

type Request struct {
	ReqURL             string
	Title              string
	InteractionChannel string
	// StreamURL string
}

func New(query, channelId string) (*Request, error) {
	var title string
	var reqUrl string
	if isValidURL(query) {
		output, err := exec.Command("youtube-dl", "-e", query).CombinedOutput()
		if err != nil {
			err = checkAgeVerification(err, string(output))
			return nil, err
		}
		title = string(output)
		reqUrl = query
	} else {
		output, err := exec.Command("youtube-dl", "-j", "ytsearch:"+query).CombinedOutput()
		if err != nil {
			err = checkAgeVerification(err, string(output))
			return nil, err
		}
		var info map[string]interface{}
		err = json.Unmarshal(output, &info)
		if err != nil {
			return nil, err
		}
		title = info["title"].(string)
		reqUrl = info["webpage_url"].(string)
	}
	return &Request{reqUrl, strings.TrimSuffix(string(title), "\n"), channelId}, nil
}

func checkAgeVerification(err error, output string) error {
	match, regexerr := regexp.MatchString("ERROR: Sign in to confirm your age", output)
	if match && regexerr == nil {
		err = errors.New("this video requires age verification")
	} else {
		err = errors.New("there was an error retrieving the video: " + err.Error())
	}
	return err
}

func (r Request) GetStreamURL() (string, error) {
	streamUrl, err := exec.Command("youtube-dl", "-f", "bestaudio", "-g", r.ReqURL).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(streamUrl), "\n"), nil
}
