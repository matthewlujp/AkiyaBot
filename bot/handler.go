package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/matthewlujp/AkiyaBot/bot/gdrive"
	"github.com/matthewlujp/AkiyaBot/photo-server/camera"
	"github.com/nlopes/slack"
)

func getCameras() ([]string, error) {
	req := fmt.Sprintf("%s/cameras", *photoServiceURL)
	res, err := http.Get(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("camera request status: %s", res.Status)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var cameras camera.Cameras
	if err := json.Unmarshal(buf, &cameras); err != nil {
		return nil, err
	}
	return cameras.IDs, nil
}

func lookupCameraDeviceName(deviceName string) string {
	// Lookup
	return deviceName
}

func getPhotos(saveDirPath string) ([]string, error) {
	var photoFilePaths []string

	camers, err := getCameras()
	if err != nil {
		return nil, err
	}

	for _, cam := range camers {
		req := fmt.Sprintf("%s/img/%s", *photoServiceURL, cam)
		logger.Printf("photo request: %s", req)
		res, err := http.Get(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("photo response status: %d", res.StatusCode)
		}

		if err = os.Mkdir(saveDirPath, 0777); err != nil && err != os.ErrExist {
			return nil, err
		}
		f, err := os.OpenFile(
			path.Join(saveDirPath, fmt.Sprintf("%s.jpg", lookupCameraDeviceName(cam))),
			os.O_RDWR|os.O_CREATE, 0777,
		)
		if err != nil {
			logger.Print(err)
			return nil, err
		}
		w := bufio.NewWriter(f)
		if _, err := io.Copy(w, res.Body); err != nil {
			logger.Print(err)
			return nil, err
		}
		if err := f.Close(); err != nil {
			logger.Print(err)
			return nil, err
		}

		photoFilePaths = append(photoFilePaths, f.Name())
	}

	return photoFilePaths, nil
}

func photoUploader(msgChannel string, rtm *slack.RTM, client *slack.Client, gService *gdrive.APIService) error {
	now := time.Now()
	baseName := now.Format("2006-01-02_15-04-05")
	dirPath := path.Join(cnf.PhotoService.SaveDir, baseName)
	fileNames, err := getPhotos(dirPath)
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("while getting photos: %v", err), msgChannel))
		return err
	}

	if len(fileNames) <= 0 {
		rtm.SendMessage(rtm.NewOutgoingMessage("写真が取れなかったよ、、、", msgChannel))
		return errors.New("no photos obtained")
	}

	for _, fileName := range fileNames {
		_, err = client.UploadFile(slack.FileUploadParameters{
			File:           fileName,
			Filetype:       "jpeg",
			Filename:       path.Base(fileName),
			Title:          "foo",
			InitialComment: "bar",
			Channels:       []string{msgChannel},
		})
		if err != nil {
			logger.Print(err)
			continue
		}

		f, err := os.Open(fileName)
		if err != nil {
			logger.Print(err)
			continue
		}
		year, month, day := now.Date()
		err = gService.Upload(&gdrive.UploadFileInfo{
			Datetime: now,
			Title:    path.Base(fileName),
			File:     f,
			Path: []string{
				"YasaiLog", strconv.Itoa(year), month.String(), fmt.Sprintf("%dth", day),
				fmt.Sprintf("%d:%d:%d", now.Hour(), now.Minute(), now.Second()),
			},
		})
		if err != nil {
			logger.Print(err)
			return err
		}
	}

	return nil
}

func (s *slackListener) handleMessageEvent(rtm *slack.RTM, ev *slack.MessageEvent, gService *gdrive.APIService) error {
	logger.Printf("MESSAGE EVENT %s:%s \"%s\"", ev.Channel, ev.User, ev.Text)

	if strings.Contains(ev.Text, "野菜の様子") {
		rtm.SendMessage(rtm.NewOutgoingMessage("野菜の写真を撮ります。", ev.Channel))
		if err := photoUploader(ev.Channel, rtm, s.client, gService); err != nil {
			logger.Print(err)
			return err
		}
	}

	return nil
}

func (s *slackListener) listenAndResponse(gService *gdrive.APIService) {
	// Start listening slack events
	rtm := s.client.NewRTM()
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(rtm, ev, gService); err != nil {
				logger.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}
