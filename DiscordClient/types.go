package DiscordClient

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type Watcher interface {
	RecentVideo(channelIds []string, timeAfter time.Time) ([]Video, error)
	ExtractIDs(urls []string) (map[string]string, error)
}

type Video interface {
	//primarily used as convenient unpack type from Watcher.RecentVideo() for later processing
	ChannelID() string
	VideoLink() string
	PublishTime() string
}

type Database interface {
	AddChannels(idMap map[string]string) error
	UpdateVideo(channelID string, lastVideo string, updateTime string) bool
	UpdateChanID(channel string, channelID string) error
	//RecentVidFromURL(channel string) (lastVideo string, updateTime string, err error)
	WatchList() []string
}

type DiscordClient struct {
	*discordgo.Session
	guild          string
	channels       []discordgo.Channel
	_events        chan bool
	APIKey         string
	commmandPrefix string
	botID          string
	PostChannel    string
	Log            *logrus.Logger
	WatchClient    Watcher
	DatabaseClient Database
}
