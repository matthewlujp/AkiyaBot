package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/blackjack/webcam"
	"github.com/labstack/echo"
)

// var cameraDevPattern = regexp.MustCompile("video[0-9]+")
var cameraDevPattern = regexp.MustCompile("ttyr([0-9]|[abcdef]+)")

type cameras struct {
	IDs []string `json:"ids"`
}

func cameraIDRequestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		devices, err := ioutil.ReadDir("/dev")
		if err != nil {
			return c.String(http.StatusInternalServerError, "cannot find devices")
		}

		var cameraIDs []string
		for _, dev := range devices {
			matched := cameraDevPattern.MatchString(dev.Name())
			if matched {
				cameraIDs = append(cameraIDs, dev.Name())
			}
		}

		return c.JSON(http.StatusOK, cameras{IDs: cameraIDs})
	}
}

func imgRequestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		cameraName := c.Param("camera_id")
		cameraFile := fmt.Sprintf("/dev/%s", cameraName)
		cam, err := webcam.Open(cameraFile) // Open webcam
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprinf("cannot open camera %s", cameraFile))
		}
		defer cam.Close()

		if err := cam.StartStreaming(); err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprinf("cannot start streaming %s", cameraFile))
		}

		// Gat frame
		timeout := uint32(10)
		err = cam.WaitForFrame(timeout)
		switch err.(type) {
		case nil:
		case *webcam.Timeout:
			return c.String(http.StatusInternalServerError, fmt.Sprinf("%s timeout", cameraFile))
		default:
			return c.String(http.StatusInternalServerError, fmt.Sprinf("%s other error while waiting for a frame", cameraFile))
		}

		frame, err := cam.ReadFrame()
		if err != nil || len(frame) <= 0 {
			return c.String(http.StatusInternalServerError, fmt.Sprinf("%s failed to get a frame", cameraFile))
		}

		return c.Blob(http.StatusOK, "image/jpeg", frame)
	}
}
