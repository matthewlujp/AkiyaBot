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

func (s *slackListener) listenAndResponse() {
	// Start listening slack events
	rtm := s.client.NewRTM()
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := handleMessageEvent(s, rtm, ev); err != nil {
				logger.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}
