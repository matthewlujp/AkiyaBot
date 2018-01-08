package photoApi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
)

// ImageFile holds info of image file
type ImageFile struct {
	Name  string // without extension
	Bytes []byte
}

// Save saves given images under a given directory
// Fils if the given dir doesn't exist
func (img *ImageFile) Save(saveDirPath string) error {
	f, err := os.OpenFile(
		path.Join(saveDirPath, fmt.Sprintf("%s.jpg", img.Name)),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766,
	)
	if err != nil {
		logger.Print(err)
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	r := bytes.NewReader(img.Bytes)
	if _, err := io.Copy(w, r); err != nil {
		logger.Printf("failed to copy from bytes to file writer, %s", err)
		return err
	}
	return nil
}
