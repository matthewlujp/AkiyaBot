package photoApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
	res, err := http.Get(fmt.Sprintf("%s/img/%s", c.serviceURL, device))
	if err != nil {
		logger.Printf("photo request to %s, %s", device, err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("photo response status: %d", res.StatusCode)
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

	images := make([]ImageFile, 0, len(cameras))
	for _, cam := range cameras {
		data, err := c.GetPhoto(cam)
		if err == nil {
			images = append(images, ImageFile{Name: lookupCameraDeviceName(cam), Bytes: data})
		}
	}
	return images, nil
}
