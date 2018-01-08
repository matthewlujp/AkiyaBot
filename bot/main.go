package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/matthewlujp/AkiyaBot/bot/gdrive"
	"github.com/matthewlujp/AkiyaBot/bot/photo-api"
	"github.com/matthewlujp/AkiyaBot/bot/watcher"
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
	Watcher      WatcherConf
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

// WatcherConf config related to regular observation
type WatcherConf struct {
	ChannelsFilePath string
}

var (
	cnf             Conf
	logger          = log.New(os.Stdout, "", log.Lshortfile)
	photoServiceURL *string
	gService        *gdrive.APIService
	apiErr          error
	photoClient     *photoApi.Client
	wtc             *watcher.Watcher
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

	wtc = &watcher.Watcher{
		ChannelsFilePath: cnf.Watcher.ChannelsFilePath,
		WatchHour:        []int{8, 12, 16, 20},
	}
}

func main() {
	slackListener := &slackListener{
		client: slack.New(cnf.Bot.BotUserOAuthAccessToken),
		botID:  cnf.Bot.ClientID,
	}

	wtc.RunPeriodic(func(w *watcher.Watcher) {
		channels, err := w.RegisteredChannels()
		if err != nil {
			logger.Printf("regular observation, get channels, %s", err)
			return
		}
		slackListener.sendMessage(channels, "regular observation!")
	}, 10*time.Second)

	slackListener.listenAndResponse()
}
