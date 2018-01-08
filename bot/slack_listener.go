package main

import (
	"io"

	"github.com/nlopes/slack"
)

type slackListener struct {
	client *slack.Client
	botID  string
}

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
