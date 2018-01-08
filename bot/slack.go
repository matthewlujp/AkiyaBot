package main

import (
	"io"

	"github.com/nlopes/slack"
)

type slackMessageSender struct {
	client *slack.Client
	rtm    *slack.RTM
}

func (sms *slackMessageSender) sendMessage(channel, message string) {
	sms.rtm.SendMessage(sms.rtm.NewOutgoingMessage(message, channel))
}

func (sms *slackMessageSender) attachData(channels []string, title string, reader io.Reader) error {
	_, err := sms.client.UploadFile(slack.FileUploadParameters{
		Reader:   reader,
		Filetype: "jpeg",
		Title:    title,
		Channels: channels,
	})
	return err
}
