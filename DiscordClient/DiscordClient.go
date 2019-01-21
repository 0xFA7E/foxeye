package DiscordClient

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

func (c *DiscordClient) Init() error {
	var err error
	c._events = make(chan bool)

	if c.commandPrefix == "" {
		c.commandPrefix = "f!"
	}

	if c.APIKey == "" {
		log.Fatalf("APIKey not set, cannot initialize")
	}
	if c.WatchClient == nil {
		return errors.New("WatchClient emplty")
	}
	if c.DatabaseClient == nil {
		return errors.New("DatabaseClient empty")
	}
	c.Session, err = discordgo.New("Bot " + c.APIKey)
	if err != nil {
		return err
	}
	//retrieve bot ID
	user, nerr := c.Session.User("@me")
	if nerr != nil {
		log.Fatalf("Could not retrieve bot ID")
	}
	c.botID = user.ID
	c.RouteMap = []Route{
		Route{
			Prefix:     "ping",
			Call:       c.ping,
			Restricted: false,
		},
		Route{
			Prefix:     "play",
			Call:       c.quickLinkPlay,
			Restricted: false,
		},
		Route{
			Prefix:     "linkadd",
			Call:       c.quickLinkAdd,
			Restricted: true,
		},
		Route{
			Prefix:     "linkremove",
			Call:       c.quickLinkRemove,
			Restricted: true,
		},
		Route{
			Prefix:     "add",
			Call:       c.addChannel,
			Restricted: true,
		},
		Route{
			Prefix:     "remove",
			Call:       c.removeChannel,
			Restricted: true,
		},
		Route{
			Prefix:     "modadd",
			Call:       c.addMod,
			Restricted: true,
		},
		Route{
			Prefix:     "modremove",
			Call:       c.removeMod,
			Restricted: true,
		},
	}
	c.AddHandler(c.commandHandler)
	c.AddHandler(c.guildCreate)
	err = c.Open()
	if err != nil {
		return err
	}
	<-c._events
	return nil
}

func (c *DiscordClient) MonitorUploads(rate time.Duration) {
	for {
		t := time.Now()
		timer := time.NewTimer(rate * time.Second)
		<-timer.C
		videoList := c.DatabaseClient.WatchList()
		videos, err := c.WatchClient.RecentVideo(videoList, t)
		if err != nil {
			c.Log.WithFields(logrus.Fields{
				"videos": videos,
				"error":  err,
			}).Error("RecentVideo pull had a problem")
		}
		if videos != nil {
			for _, v := range videos {
				if c.DatabaseClient.UpdateVideo(v.ChannelID(), v.VideoLink(), v.PublishTime()) {
					err := c.SendByName(c.PostChannel, v.VideoLink())
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
}

func (c *DiscordClient) commandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	//ignore msg created by bot
	user := m.Author
	if user.ID == c.botID || user.Bot {
		return
	}
	if strings.HasPrefix(m.Content, c.commandPrefix) {
		split := strings.Split(m.Content, c.commandPrefix)
		split = strings.Split(split[1], " ")
		for _, route := range c.RouteMap {
			if split[0] == route.Prefix {
				if route.Restricted == true {
					authorized := c.userAuthorized(s, m, user.ID)
					if authorized == true {
						route.Call(s, m, split[1:])
					} else {
						return
					}
				} else {
					route.Call(s, m, split[1:])
				}
			}
		}
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
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("No chan name found by name: " + chanName)

}

func (c *DiscordClient) reply(m *discordgo.MessageCreate, response string) {
	_, err := c.ChannelMessageSend(m.ChannelID, m.Author.Mention()+response)
	if err != nil {
		c.Log.Errorln(err)
	}
}

func (c *DiscordClient) userAuthorized(s *discordgo.Session, m *discordgo.MessageCreate, userID string) bool {
	//first we pull the users roles and check if theyre an admin, that overrides permission, then we check the db
	member, err := s.State.Member(m.GuildID, userID)
	if err != nil {
		c.Log.Errorf("Failed getting member info:%v", err)
		return false
	}
	for _, roleID := range member.Roles {
		role, err := s.State.Role(m.GuildID, roleID)
		if err != nil {
			c.Log.Error(err)
			return false
		}
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true
		}
	}

	//user is not an admin, check db for authorized users
	authorized := c.DatabaseClient.IsAuthorized(userID)
	fmt.Println(authorized)
	return authorized
}
