package DiscordClient

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/0xFA7E/foxeye/SqliteClient"
	"github.com/sirupsen/logrus"

	"github.com/0xFA7E/foxeye/YTClient"
	"github.com/bwmarrin/discordgo"
)

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
	YTClient       *YTClient.YoutubeClient
	SQLCli         *SqliteClient.SQLCli
}

func (c *DiscordClient) Init() {
	var err error
	c._events = make(chan bool)

	if c.commmandPrefix == "" {
		c.commmandPrefix = "f!"
	}

	if c.APIKey == "" {
		log.Fatalf("APIKey not set, cannot initialize")
	}
	c.Session, err = discordgo.New("Bot " + c.APIKey)
	if err != nil {
		log.Fatalf("Error creating session: %v", err)
		return
	}
	//retrieve bot ID
	user, nerr := c.Session.User("@me")
	if nerr != nil {
		log.Fatalf("Could not retrieve bot ID")
	}
	c.botID = user.ID
	c.AddHandler(c.commandHandler)
	c.AddHandler(c.guildCreate)
	err = c.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
		return
	}

	<-c._events
	//fmt.Println("Popped")
	return
}

func (c *DiscordClient) MonitorUploads(rate time.Duration) {
	for {
		t := time.Now()
		timer := time.NewTimer(rate * time.Second)
		<-timer.C
		videos := c.YTClient.RecentVideo(t)
		if videos != nil {
			for _, v := range videos {
				err := c.SendByName(c.PostChannel, v)
				if err != nil {
					c.Log.WithFields(logrus.Fields{
						"Post Channel": c.PostChannel,
						"Video":        v,
					}).Error("Error sending message")
				}
			}
		}
	}
}

func (c *DiscordClient) commandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	//ignore msg created by bot
	user := m.Author
	if user.ID == c.botID || user.Bot {
		return
	}
	c.ping(s, m, c.commmandPrefix)
	c.addChannel(s, m, c.commmandPrefix)
}

func (c *DiscordClient) addChannel(s *discordgo.Session, m *discordgo.MessageCreate, prefix string) {
	if strings.HasPrefix(m.Content, prefix+"watch") {
		split := strings.Split(m.Content, " ")
		//we cutoff the first part of the split, as its the prefix.
		err := c.YTClient.AddChannels(split[1:])
		if err != nil {
			log.Printf("Failed to add channels:%v\n", err)
			c.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Failed to add channels. See logs for more details")
		} else {
			log.Println("Added new channels to DB")
			c.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Added new channels!")
		}
	}
}

func (c *DiscordClient) ping(s *discordgo.Session, m *discordgo.MessageCreate, prefix string) {
	if strings.HasPrefix(m.Content, prefix+"ping") {
		c.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" Pong!")
	}
}

func (c *DiscordClient) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	c.guild = event.Guild.ID
	for _, channel := range event.Guild.Channels {
		fmt.Println("Found channel: ", channel.Name, " Type: ", channel.Type)
		c.channels = append(c.channels, *channel)

	}
	// let context know we received an event
	//We do this because the guild create event may come a tad slower than we'd like
	//and dont want to accidentally try to do some work before were ready
	c.SendByName(c.PostChannel, "Foxeye is on the watch!")
	c._events <- true
	return
}

//sends a message by a given channel. *WARNING* would send message to first channels if same name!
func (c *DiscordClient) SendByName(chanName string, msg string) error {
	for _, channel := range c.channels {
		if channel.Name == chanName {
			_, err := c.ChannelMessageSend(channel.ID, msg)
			//fmt.Println(s)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("No chan name found by name: " + chanName)
}
