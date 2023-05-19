package request

import (
	"regexp"
)

// func isValidURL(reqUrl string) bool {
// 	// Check it's an Absolute URL or absolute path
// 	uri, err := url.Parse(reqUrl)
// 	if err != nil {
// 		return false
// 	}

// 	// Check it's a valid domain name
// 	_, err = net.LookupHost(uri.Host)
// 	return err == nil
// }

type Request struct {
	RequestURL         string
	Title              string
	InteractionChannel string
	// StreamURL string
}

// func New(query, channelId string) (*Request, error) {
// 	var title string
// 	var reqUrl string
// 	if isValidURL(query) {
// 		output, err := exec.Command("yt-dlp", "-e", query).CombinedOutput()
// 		if err != nil {
// 			err = checkAgeVerification(err, string(output))
// 			return nil, err
// 		}
// 		title = string(output)
// 		reqUrl = query
// 	} else {
// 		output, err := exec.Command("yt-dlp", "-j", "ytsearch:"+query).CombinedOutput()
// 		if err != nil {
// 			err = checkAgeVerification(err, string(output))
// 			return nil, err
// 		}
// 		var info map[string]interface{}
// 		err = json.Unmarshal(output, &info)
// 		if err != nil {
// 			return nil, err
// 		}
// 		title = info["title"].(string)
// 		reqUrl = info["webpage_url"].(string)
// 	}
// 	return &Request{reqUrl, strings.TrimSuffix(string(title), "\n"), channelId}, nil
// }

func Generate(query, channelId string, songRequests chan *Request) {
	var cmd string = "youtube-dl"
	var title, reqUrl string
	switch {
	case regexp.MustCompile("youtube.com/playlist").MatchString(query):
		cmdOpts := []string{"--flat-playlist", "-J", query}
		info := ytdlCmdExec(cmd, cmdOpts)
		entries := (info["entries"]).([]interface{})
		for _, entry := range entries {
			newQuery := entry.(map[string]interface{})["url"].(string)
			Generate(newQuery, channelId, songRequests)
		}
		return
	case regexp.MustCompile("youtube.com/watch").MatchString(query):
		cmdOpts := []string{"--no-playlist", "-J", query}
		info := ytdlCmdExec(cmd, cmdOpts)
		title = info["title"].(string)
		reqUrl = info["webpage_url"].(string)
	default:
		cmdOpts := []string{"--no-playlist", "-J", "ytsearch:" + query}
		info := ytdlCmdExec(cmd, cmdOpts)
		entry := (info["entries"]).([]interface{})[0]
		title = entry.(map[string]interface{})["title"].(string)
		reqUrl = entry.(map[string]interface{})["webpage_url"].(string)
	}
	songRequest := &Request{reqUrl, title, channelId}
	songRequests <- songRequest
}

func (r Request) GetStreamURL() (string, error) {
	streamUrl, err := ytdlGetStream(r.RequestURL)
	if err != nil {
		return "", err
	}
	return streamUrl, nil
}

// func checkAgeVerification(err error, output string) error {
// 	match, regexerr := regexp.MatchString("ERROR: Sign in to confirm your age", output)
// 	if match && regexerr == nil {
// 		err = errors.New("this video requires age verification")
// 	} else {
// 		err = errors.New("there was an error retrieving the video: " + err.Error())
// 	}
// 	return err
// }

// defunct - see ytdl.go
// func (r Request) GetStreamURL() (string, error) {
// 	streamUrl, err := exec.Command("yt-dlp", "-f", "bestaudio", "-g", r.ReqURL).Output()
// 	if err != nil {
// 		return "", err
// 	}
// 	return strings.TrimSuffix(string(streamUrl), "\n"), nil
// }
