package main

import (
	"flag"
	"log"
	"os"

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
	BotID                   string
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

func listenAndResponse(s *slackListener) {
	// Start listening slack events
	rtm := s.client.NewRTM()
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := handleMessageEvent(s, rtm, ev); err != nil {
				logger.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

func main() {
	slackListener := &slackListener{
		client: slack.New(cnf.Bot.BotUserOAuthAccessToken),
		botID:  cnf.Bot.ClientID,
	}

	wtc.Run(func(w *watcher.Watcher) error {
		channels, err := w.RegisteredChannels()
		if err != nil {
			logger.Printf("regular observation get channels failed, %s", err)
			return err
		}

		slackListener.sendMessage(channels, "定期観察だよ")
		for _, ch := range channels {
			if err := takePhotoAndProcess(ch, slackListener); err != nil {
				logger.Printf("regular observation channel %s, %s", ch, err)
				return err
			}
		}
		return nil
	})

	listenAndResponse(slackListener)
}
