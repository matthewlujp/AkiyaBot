package watcher

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func (wtc *Watcher) readChannelsFromFile() ([]string, error) {
	f, err := os.Open(wtc.ChannelsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		logger.Print(err)
		return nil, err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		logger.Print(err)
		return nil, err
	}
	return strings.Split(string(buf), "\n"), nil
}

// RegisteredChannels returns channels registered for regular observation
func (wtc *Watcher) RegisteredChannels() ([]string, error) {
	return wtc.readChannelsFromFile()
}

// RegisterChannel registers new channel
func (wtc *Watcher) RegisterChannel(channelID string) error {
	regisered, err := wtc.IsRegistered(channelID)
	if err != nil {
		return err
	}
	if regisered {
		return nil
	}

	f, err := os.OpenFile(wtc.ChannelsFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		logger.Printf("register channel open %s, %s", wtc.ChannelsFilePath, err)
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("%s\n", channelID)); err != nil {
		logger.Print(err)
		return err
	}

	return nil
}

// IsRegistered tells whether a given channel is registered for regular observation
func (wtc *Watcher) IsRegistered(channelID string) (bool, error) {
	registeredChannels, err := wtc.readChannelsFromFile()
	if err != nil {
		logger.Print(err)
		return false, err
	}
	for _, ch := range registeredChannels {
		if ch == channelID {
			return true, nil
		}
	}
	return false, nil
}

// DeregisterChannel remove a given channel from registered channels
func (wtc *Watcher) DeregisterChannel(channelID string) error {
	registeredChannels, err := wtc.readChannelsFromFile()
	if err != nil {
		logger.Print(err)
		return err
	}
	logger.Printf("registerd ids: [%s], deregister %s", strings.Join(registeredChannels, ", "), channelID)

	registered := false
	for _, ch := range registeredChannels {
		if ch == channelID {
			registered = true
			break
		}
	}
	if !registered {
		return nil
	}

	f, err := os.OpenFile(wtc.ChannelsFilePath, os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		logger.Print(err)
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, ch := range registeredChannels {
		if ch != channelID {
			logger.Printf("rewrite %s", ch)
			if _, err = w.WriteString(fmt.Sprintf("%s\n", ch)); err != nil {
				logger.Print(err)
				return err
			}
		}
	}
	w.Flush()
	return nil
}
