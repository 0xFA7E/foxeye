package DiscordClient

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type ServerContext struct {
	*discordgo.Session
	guild    string
	channels []discordgo.Channel
	_events  chan bool
	APIKey   string
}

func (c *ServerContext) Init() {
	var err error
	c._events = make(chan bool)
	if c.APIKey == "" {
		log.Fatalf("APIKey not set, cannot initialize")
	}
	c.Session, err = discordgo.New("Bot " + c.APIKey)
	if err != nil {
		log.Fatalf("Error creating session: %v", err)
		return
	}
	c.AddHandler(c.messageCreate)
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

func (c *ServerContext) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	//ignore msg created by bot
	if m.Author.ID == s.State.User.ID {
		return
	}
	//if message is "ping" reply with "pong"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
	//if msg is pong give ping
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func (c *ServerContext) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
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
	c._events <- true
	//fmt.Println("sent")
	return
}

//sends a message by a given channel. *WARNING* would send message to first channels if same name!
func (c *ServerContext) SendByName(chanName string, msg string) error {
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
