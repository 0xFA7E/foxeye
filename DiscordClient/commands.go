package DiscordClient

import (
	"github.com/bwmarrin/discordgo"
)

func (c *DiscordClient) addMod(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		c.Log.Println("Too few args for adding mod")
		c.reply(m, " Too few arguments for adding a mod")
	}
	username := args[0]
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		c.Log.Errorf("Error getting guild ID: %v", err)
		c.reply(m, " Error adding mod, see log for more details.")
		return
	}
	for _, member := range guild.Members {
		userTag := member.User.Username + "#" + member.User.Discriminator
		if userTag == username {
			err = c.DatabaseClient.AddMod(member.User.ID)
			if err != nil {
				c.Log.Errorf("Error adding moderator: %v", err)
				c.reply(m, " Error adding mod, see log for more details.")
				return
			} else {
				c.reply(m, " Added mod")
				return
			}
		}
	}
	c.reply(m, " No user found to add")
}

func (c *DiscordClient) removeMod(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		c.Log.Printf("Too few arguments to remove mod\n")
		c.reply(m, " Failed to remove mod, not enough arguments")
		return
	}
	username := args[0]
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		c.Log.Errorf("Error getting guild ID: %v", err)
		c.reply(m, " Error removing mod, see log for more details.")
	}
	for _, member := range guild.Members {
		userTag := member.User.Username + "#" + member.User.Discriminator
		if userTag == username {
			err = c.DatabaseClient.RemoveMod(member.User.ID)
			if err != nil {
				c.Log.Errorf("Error removing moderator: %v", err)
				c.reply(m, " Error removing mod, see log for more details.")
				return
			} else {
				c.reply(m, " Removed mod")
				return
			}
		}
	}
	c.reply(m, " No user found to remove")
}

func (c *DiscordClient) quickLinkPlay(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		c.Log.Printf("Too few arguments for link\n")
		c.reply(m, " Failed to play link not enough arguments")
		return
	}
	linkID := args[0]
	link, err := c.DatabaseClient.FetchLink(linkID)
	if err != nil {
		c.Log.Printf("Failed to fetch link: %v\n", err)
		c.reply(m, " Failed to fetch link, see log for more details")
		return
	}
	if link != "" {
		c.ChannelMessageSend(m.ChannelID, link)
	} else {
		c.Log.Printf("Link id not found:%v", linkID)
		c.reply(m, " Not a valid link")
	}
}

func (c *DiscordClient) quickLinkAdd(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 2 {
		c.Log.Printf("Too few arguments for quick link add\n")
		c.reply(m, " Failed to add a new quick link, not enough arguments")
		return
	}
	linkID := args[0]
	link := args[1]
	err := c.DatabaseClient.AddLink(linkID, link)
	if err != nil {
		c.Log.Printf("Failed to add link: %v\n", err)
		c.reply(m, " Failed to add a new quick link, see log for more details")
	}
	c.reply(m, " Added new link")
}

func (c *DiscordClient) quickLinkRemove(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) < 1 {
		c.Log.Printf("Too few arguments for quick link remove\n")
		c.reply(m, " Failed to remove quick link, not enough arguments")
		return
	}
	linkID := args[0]
	err := c.DatabaseClient.RemoveLink(linkID)
	if err != nil {
		c.Log.Printf("Failed to remove link: %v\n", err)
		c.reply(m, " Failed to remove quick link, see log for more details")
	}
	c.reply(m, " Removed Link.")
}

func (c *DiscordClient) addChannel(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) <= 0 {
		c.Log.Printf("No channels given to add")
		c.reply(m, " No channels given!")
		return
	}
	ids, err := c.WatchClient.ExtractIDs(args)
	if err != nil {
		c.Log.Printf("Failed to add channels:%v\n", err)
		c.reply(m, " Failed to add channels. See logs for more details")
	} else {
		err = c.DatabaseClient.AddChannels(ids)
		if err != nil {
			c.Log.Printf("Failed to add channels:%v\n", err)
		} else {
			c.Log.Println("Added new channels to DB")
			c.reply(m, " Added new channels!")
		}
	}
}

func (c *DiscordClient) removeChannel(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) <= 0 {
		c.Log.Printf("Could not remove channels, none given")
		c.reply(m, "Cant remove channels you dont tell me about!")
	}

	ids, err := c.WatchClient.ExtractIDs(args)
	if err != nil {
		c.Log.Printf("Failed to remove channels%v\n", err)
		c.reply(m, " Failed to remove channels. See logs for more details")
	} else {
		err = c.DatabaseClient.RemoveChannels(ids)
		if err != nil {
			c.Log.Printf("Failed to remove channels%v\n", err)
			c.reply(m, " Failed to remove channels. See logs for more details")
		} else {
			c.Log.Println("Removed channels from DB")
			c.reply(m, " Removed channels!")
		}
	}

}

func (c *DiscordClient) ping(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	/*exists as an easy test for command functions*/
	c.reply(m, " Pong!")
}
