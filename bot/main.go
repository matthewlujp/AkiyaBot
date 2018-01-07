package main

import (
	"flag"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/matthewlujp/AkiyaBot/bot/gdrive"
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

type slackListener struct {
	client    *slack.Client
	botID     string
	channelID string
}

var (
	cnf             Conf
	logger          = log.New(os.Stdout, "", log.Lshortfile)
	photoServiceURL *string
	gService        *gdrive.APIService
	apiErr          error
)

func init() {
	if _, err := toml.DecodeFile(configFile, &cnf); err != nil {
		logger.Fatalf("error while parsing conf toml: %s", err)
	}

	photoServiceURL = flag.String("photo_url", cnf.PhotoService.URL, "port number for forwarding")
	flag.Parse()

	gService, apiErr = gdrive.GetAPIService(cnf.GDriveAPI.ClientSecretPath)
	if apiErr != nil {
		logger.Fatalf("failed to obtain gdrive api service %s", apiErr)
	}
}

func main() {
	client := slack.New(cnf.Bot.BotUserOAuthAccessToken)
	slackListener := &slackListener{
		client:    client,
		botID:     cnf.Bot.ClientID,
		channelID: "test",
	}
	slackListener.listenAndResponse(gService)
}
