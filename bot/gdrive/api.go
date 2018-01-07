package gdrive

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

const folderMime = "application/vnd.google-apps.folder"

// UploadFileInfo holds information and reader for a file to be upoaded
type UploadFileInfo struct {
	Datetime time.Time
	Title    string
	File     io.Reader
	Path     Path
}

// APIService is alias for drive.Service
type APIService drive.Service

// Path is a list of folders from the root
type Path []string

var (
	logger = log.New(os.Stdout, "", log.Lshortfile)
)

// GetAPIService returns APIService instance
func GetAPIService(clientSecretPath string) (*APIService, error) {
	b, err := ioutil.ReadFile(clientSecretPath)
	if err != nil {
		logger.Printf("Unable to read client secret file: %v", err)
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		logger.Printf("Unable to parse client secret file to config: %v", err)
		return nil, err
	}
	client, err := getClient(context.Background(), config)
	if err != nil {
		return nil, err
	}
	srv, err := drive.New(client)
	if err != nil {
		return nil, err
	}
	return (*APIService)(srv), nil
}

// Upload uploads a given file under specified paraent folder
func (gService *APIService) Upload(upFile *UploadFileInfo) error {
	folder, err := gService.createPath(upFile.Path)
	if err != nil {
		return err
	}

	_, err = gService.Files.Create(&drive.File{Name: upFile.Title, Parents: []string{folder.Id}}).Media(upFile.File).Do()
	return err
}

func (gService *APIService) createPath(path Path) (*drive.File, error) {
	parentFolder := &drive.File{
		MimeType: folderMime,
	}

PATH_SEARCH:
	for _, folderName := range path {
		entities, err := gService.listFolder(parentFolder.Id)
		if err != nil {
			logger.Print(err)
			return nil, err
		}
		for _, e := range entities {
			if e.Name == folderName {
				parentFolder.Name = e.Name
				parentFolder.Id = e.Id
				continue PATH_SEARCH
			}
		}
		// target folderName not exists
		newFolder, err := gService.createFolder(folderName, parentFolder.Id)
		if err != nil {
			logger.Print(err)
			return nil, err
		}
		parentFolder.Name = newFolder.Name
		parentFolder.Id = newFolder.Id
	}

	return parentFolder, nil
}

func (gService *APIService) createFolder(name, parentID string) (*drive.File, error) {
	f := &drive.File{Name: name, MimeType: folderMime}
	if parentID != "" {
		f.Parents = []string{parentID}
	}
	return gService.Files.Create(f).Do()
}

// If list target is root, parent should be ""
func (gService *APIService) listFolder(parentID string) ([]*drive.File, error) {
	query := fmt.Sprintf("not trashed and mimeType = '%s'", folderMime)
	if parentID != "" {
		query += fmt.Sprintf(" and '%s' in parents", parentID)
	}
	logger.Printf("query folder list: %s", query)

	list, err := gService.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		logger.Print(err)
		return nil, err
	}
	return list.Files, nil
}
