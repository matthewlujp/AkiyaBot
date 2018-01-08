package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/matthewlujp/AkiyaBot/bot/gdrive"
	"github.com/matthewlujp/AkiyaBot/bot/photo-api"
	"github.com/nlopes/slack"
)

func takePhotoAndProcess(channel string, sms *slackMessageSender) error {
	imageFiles, err := photoClient.GetAllPhotos()
	if err != nil {
		return err
	}

	if len(imageFiles) <= 0 {
		sms.sendMessage(channel, "写真が撮れなかったよ")
		return errors.New("no photos obtained")
	}

	wg := &sync.WaitGroup{}

	// Attach image files to slack
	for _, f := range imageFiles {
		wg.Add(1)
		go func(imgFile *photoApi.ImageFile) {
			if err = sms.attachData([]string{channel}, imgFile.Name, bytes.NewReader(imgFile.Bytes)); err != nil {
				logger.Printf("failed to attach %s, %s", imgFile.Name, err)
			}
			wg.Done()
		}(&f)
	}

	now := time.Now()

	// Upload files to google drive
	for _, f := range imageFiles {
		wg.Add(1)
		go func(imgFile *photoApi.ImageFile) {
			err = gService.Upload(&gdrive.UploadFileInfo{
				Datetime: now,
				Title:    imgFile.Name,
				Reader:   bytes.NewReader(imgFile.Bytes),
				Path:     pathFromDateTime(now),
			})
			if err != nil {
				logger.Printf("upload %s to google drive, %s", imgFile.Name, err)
			}
		}(&f)
	}

	// Save on local
	saveDirPath := saveDirFromTime(now)
	if err = os.MkdirAll(saveDirPath, 0777); err != nil && err != os.ErrExist {
		logger.Printf("failed to create dir to save, %s", err)
		return err
	}
	for _, f := range imageFiles {
		wg.Add(1)
		go func(imgFile *photoApi.ImageFile) {
			imgFile.Save(saveDirPath)
			wg.Done()
		}(&f)
	}

	wg.Wait()
	return nil
}

func (s *slackListener) handleMessageEvent(rtm *slack.RTM, ev *slack.MessageEvent) error {
	logger.Printf("MESSAGE EVENT %s:%s \"%s\"", ev.Channel, ev.User, ev.Text)
	if strings.Contains(ev.Text, "野菜の様子") {
		rtm.SendMessage(rtm.NewOutgoingMessage("野菜の写真を撮ります。", ev.Channel))
		if err := takePhotoAndProcess(ev.Channel, &slackMessageSender{rtm: rtm, client: s.client}); err != nil {
			return err
		}
	}
	return nil
}

func (s *slackListener) listenAndResponse() {
	// Start listening slack events
	rtm := s.client.NewRTM()
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(rtm, ev); err != nil {
				logger.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}
