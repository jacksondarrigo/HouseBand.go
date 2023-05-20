package request

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
)

func ytdlCmdExec(cmd string, cmdOpts []string) map[string]interface{} {
	command := exec.Command(cmd, cmdOpts...)
	output, err := command.Output()
	if err != nil {
		log.Println("Playlist command error:", err.Error())
	}
	var info map[string]interface{}
	err = json.Unmarshal([]byte(output), &info)
	if err != nil {
		log.Println("JSON error:", err.Error())
	}
	return info
}

func ytdlGetStream(RequestURL string) (string, error) {
	streamUrlCmd := exec.Command("yt-dlp", "--no-playlist", "-f", "bestaudio", "-g", RequestURL)
	streamUrlOutput, err := streamUrlCmd.Output()
	if err != nil {
		return "", err
	}
	results := strings.TrimSuffix(string(streamUrlOutput), "\n")
	return results, nil
}
