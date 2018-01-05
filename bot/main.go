package main

import (
	"flag"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/nlopes/slack"
)

const (
	configFile = "./conf.toml"
)

// Conf all config info
type Conf struct {
	Bot          BotConf
	PhotoService PhotoServiceConf
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

type slackListener struct {
	client    *slack.Client
	botID     string
	channelID string
}

var (
	cnf             Conf
	logger          = log.New(os.Stdout, "", log.Lshortfile)
	photoServiceURL *string
)

func init() {
	if _, err := toml.DecodeFile(configFile, &cnf); err != nil {
		logger.Fatalf("error while parsing conf toml: %s", err)
	}

	photoServiceURL = flag.String("photo_url", cnf.PhotoService.URL, "port number for forwarding")
	flag.Parse()
}

func (s *slackListener) listenAndResponse() {
	// Start listening slack events
	rtm := s.client.NewRTM()
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(rtm, ev); err != nil {
				logger.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

func main() {
	client := slack.New(cnf.Bot.BotUserOAuthAccessToken)
	slackListener := &slackListener{
		client:    client,
		botID:     cnf.Bot.ClientID,
		channelID: "hoge",
	}

	slackListener.listenAndResponse()
}
