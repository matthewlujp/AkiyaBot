package main

import (
	"C"
	"bytes"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/labstack/echo"
	"github.com/yuntan/go-v4l2/capture"
)

var cameraDevPattern = regexp.MustCompile("video[0-9]+")

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

		return c.JSON(http.StatusOK, Cameras{IDs: cameraIDs})
	}
}

func imgRequestHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		cameraName := c.Param("camera_id")
		cameraDeviceFile := fmt.Sprintf("/dev/%s", cameraName)
		img, err := capture.Capture(cameraDeviceFile)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("%s failed to get image", cameraDeviceFile))
		}

		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, img, nil); err != nil {
			return c.String(http.StatusInternalServerError, "failed to convert a frame into jpeg")
		}
		return c.Blob(http.StatusOK, "image/jpeg", buf.Bytes())
	}
}
