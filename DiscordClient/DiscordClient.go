package DiscordClient

import (
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type ServerContext struct {
	*discordgo.Session
	guild          string
	channels       []discordgo.Channel
	_events        chan bool
	APIKey         string
	commmandPrefix string
	botID          string
}

func (c *ServerContext) Init() {
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

func (c *ServerContext) commandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	//ignore msg created by bot
	user := m.Author
	if user.ID == c.botID || user.Bot {
		return
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
