package photoApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/matthewlujp/AkiyaBot/photo-server/camera"
)

var (
	logger = log.New(os.Stdout, "[photo-api]", log.Lshortfile)
)

// Client is client object for photo api call
type Client struct {
	serviceURL string
}

// GetAPIClient returns PhotoAPIClient instance
func GetAPIClient(serviceURL string) *Client {
	return &Client{serviceURL: serviceURL}
}

// GetCameras returns device names available
func (c *Client) GetCameras() ([]string, error) {
	res, err := http.Get(fmt.Sprintf("%s/cameras", c.serviceURL))
	if err != nil {
		logger.Printf("while camera request, %s", err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		logger.Printf("camera response, %s", err)
		return nil, fmt.Errorf("camera response status: %s", res.Status)
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Print(err)
		return nil, err
	}

	var cameras camera.Cameras
	if err := json.Unmarshal(buf, &cameras); err != nil {
		logger.Print(err)
		return nil, err
	}
	return cameras.IDs, nil
}

func lookupCameraDeviceName(deviceName string) string {
	// Lookup
	return deviceName
}

// GetPhoto returns ReadCloser of retreived image
func (c *Client) GetPhoto(device string) ([]byte, error) {
	req := fmt.Sprintf("%s/img/%s", c.serviceURL, device)
	logger.Printf("photo request %s", req)
	res, err := http.Get(req)
	if err != nil {
		logger.Printf("photo request to %s, %s", device, err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var resStr string
		if buf, readErr := ioutil.ReadAll(res.Body); readErr == nil {
			resStr = string(buf)
		}
		logger.Printf("response to request %s: %s, %s", device, res.Status, resStr)
		return nil, fmt.Errorf("photo response (status %d) %s", res.StatusCode, resStr)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Printf("reading bytes from %s photo response, %s", device, err)
		return nil, err
	}
	return data, nil
}

// GetAllPhotos request server to take photos and returns a map {name: ReadCloser}
func (c *Client) GetAllPhotos() ([]ImageFile, error) {
	cameras, err := c.GetCameras()
	if err != nil {
		return nil, err
	}
	logger.Printf("all cameras %s", strings.Join(cameras, " "))

	images := make([]ImageFile, 0, len(cameras))
	for _, cam := range cameras {
		data, err := c.GetPhoto(cam)
		if err == nil {
			images = append(images, ImageFile{Name: lookupCameraDeviceName(cam), Bytes: data})
		}
	}
	return images, nil
}
