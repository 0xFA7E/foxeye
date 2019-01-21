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
	RemoveChannels(idMap map[string]string) error
	UpdateVideo(channelID string, lastVideo string, updateTime string) bool
	UpdateChanID(channel string, channelID string) error
	AddLink(linkID string, link string) error
	RemoveLink(linkID string) error
	FetchLink(linkID string) (string, error)
	IsAuthorized(userID string) bool
	WatchList() []string
	AddMod(userID string) error
	RemoveMod(userID string) error
}

type DiscordClient struct {
	*discordgo.Session
	guild          string
	channels       []discordgo.Channel
	_events        chan bool
	APIKey         string
	commandPrefix  string
	botID          string
	PostChannel    string
	Log            *logrus.Logger
	WatchClient    Watcher
	DatabaseClient Database
	RouteMap       []Route
}

type Route struct {
	Prefix     string
	Restricted bool
	Call       func(s *discordgo.Session, m *discordgo.MessageCreate, args []string)
}
