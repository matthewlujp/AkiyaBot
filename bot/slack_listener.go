package main

import (
	"io"
	"regexp"

	"github.com/nlopes/slack"
)

type slackListener struct {
	client *slack.Client
	botID  string
}

var (
	mentionPattern = regexp.MustCompile(`^<@([A-Z0-9]+)>`)
)

func (s *slackListener) sendMessage(channels []string, message string) {
	for _, ch := range channels {
		s.client.SendMessage(ch, slack.MsgOptionText(message, false))
	}
}

func (s *slackListener) attachData(channels []string, title string, reader io.Reader) error {
	_, err := s.client.UploadFile(slack.FileUploadParameters{
		Reader:   reader,
		Filetype: "jpeg",
		Title:    title,
		Filename: title,
		Channels: channels,
	})
	return err
}

// parseMention check mention tag in a text and return target user if any
func parseMention(text string) (bool, string) {
	matched := mentionPattern.FindStringSubmatch(text)
	if len(matched) == 0 {
		return false, ""
	}
	return true, matched[1]
}

func isMentioned(text, userID string) bool {
	mentioned, target := parseMention(text)
	if !mentioned {
		return false
	}
	if target == userID {
		return true
	}
	return false
}
