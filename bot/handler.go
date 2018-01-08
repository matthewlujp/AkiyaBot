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

func takePhotoAndProcess(channel string, s *slackListener) error {
	imageFiles, err := photoClient.GetAllPhotos()
	if err != nil {
		return err
	}

	if len(imageFiles) <= 0 {
		s.sendMessage([]string{channel}, "写真が撮れなかったよ")
		return errors.New("no photos obtained")
	}

	wg := &sync.WaitGroup{}

	// Attach image files to slack
	for _, f := range imageFiles {
		wg.Add(1)
		go func(imgFile *photoApi.ImageFile) {
			if err = s.attachData([]string{channel}, imgFile.Name, bytes.NewReader(imgFile.Bytes)); err != nil {
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
			wg.Done()
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

func watcherRegistration(text, channel string, s *slackListener) error {
	if strings.Contains(text, "定期観察の依頼") {
		logger.Printf("register %s as a regular observation channel", channel)
		s.sendMessage([]string{channel}, "<!here> このチャンネルに定期的に野菜の画像をアップするよ")
		wtc.RegisterChannel(channel)
	} else if strings.Contains(text, "定期観察の解除") {
		logger.Printf("deregister %s as a regular observation channel", channel)
		s.sendMessage([]string{channel}, "<!here> このチャンネルへの野菜画像のアップをやめるよ")
		wtc.DeregisterChannel(channel)
	}
	return nil
}

func handleMessageEvent(s *slackListener, rtm *slack.RTM, ev *slack.MessageEvent) error {
	logger.Printf("MESSAGE EVENT %s:%s \"%s\"", ev.Channel, ev.User, ev.Text)
	if isMentioned(ev.Text, cnf.Bot.BotID) && strings.Contains(ev.Text, "野菜の様子") {
		s.sendMessage([]string{ev.Channel}, "<!here> 野菜の写真を撮るよ")
		rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))
		if err := takePhotoAndProcess(ev.Channel, s); err != nil {
			return err
		}
	} else if isMentioned(ev.Text, cnf.Bot.BotID) && strings.Contains(ev.Text, "定期観察") {
		if err := watcherRegistration(ev.Text, ev.Channel, s); err != nil {
			return err
		}
	}
	return nil
}
