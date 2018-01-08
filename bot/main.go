package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/matthewlujp/AkiyaBot/bot/gdrive"
	"github.com/matthewlujp/AkiyaBot/bot/photo-api"
	"github.com/nlopes/slack"
)

const (
	configFile = "./conf.toml"
)

// Conf all config info
type Conf struct {
	Bot          BotConf
	PhotoService PhotoServiceConf
	GDriveAPI    GDriveAPIConf
}

// BotConf config related to bot
type BotConf struct {
	BotName                 string
	ClientID                string
	BotUserOAuthAccessToken string
}

// PhotoServiceConf config related to photo service
type PhotoServiceConf struct {
	URL     string
	SaveDir string
}

// GDriveAPIConf config related to google drive upload
type GDriveAPIConf struct {
	ClientSecretPath string
}

var (
	cnf             Conf
	logger          = log.New(os.Stdout, "", log.Lshortfile)
	photoServiceURL *string
	gService        *gdrive.APIService
	apiErr          error
	photoClient     *photoApi.Client
)

func init() {
	if _, err := toml.DecodeFile(configFile, &cnf); err != nil {
		logger.Fatalf("error while parsing conf toml: %s", err)
	}

	photoServiceURL = flag.String("photo_url", cnf.PhotoService.URL, "port number for forwarding")
	flag.Parse()

	photoClient = photoApi.GetAPIClient(*photoServiceURL)

	gService, apiErr = gdrive.GetAPIService(cnf.GDriveAPI.ClientSecretPath)
	if apiErr != nil {
		logger.Fatalf("failed to obtain gdrive api service %s", apiErr)
	}
}

func main() {
	slackListener := &slackListener{
		client: slack.New(cnf.Bot.BotUserOAuthAccessToken),
		botID:  cnf.Bot.ClientID,
	}
	go slackListener.listenAndResponse()

	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			slackListener.sendMessage([]string{"C8HT1A96V"}, "定期観測")
		}
	}
}
