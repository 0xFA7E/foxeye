# foxeye
Discord bot for monitoring youtube uploads

#Usage

foxeye -gen <configuration name>
  Youtube API Key:<YoutubeAPI Key>

  Discord API Key:<Discord API Key

  Channel to post in:<posting channel>

  sqlite DB file to use:<database filename>
  
  foxeye -config <configuration file>
  
  #Commands
  --current prefix for commands is 'f!' //TODO Add custom prefix to configuration
  With the exception of the f!play command, all commands require the user to either have an administrator role, or be added as a bot moderator.
  (note: this means bot mods can add other bot mods! Finer grain permissions in todo list, but isnt currently high priority as biggest risk is rogue moderator griefing)
  
  
  f!modadd <username#tag>
  --Adds a specified user to approved bot moderators. 
  
  f!modremove <username#tag>
  --Removes a specified user from the list of bot moderators.
  
  f!add <space seperated list of channel urls>
  --Adds listed channels to the database for monitoring uploads.
  
  f!remove <space separated list of channel urls>
  --Removes listed channels from the database.
  
  f!linkadd <linkid> <link>
  --Adds the specified <link> to be called anytime f!play <linkid> is called. Note: doesnt actually check if the link is a proper url. Technically it can link any message given to it. Get creative. 
  
  f!linkremove <linkid>
  --removes the <linkid> from the database
  
  f!play <linkid>
  --Makes the bot post the associated link(or message) corresponding to <linkid>. Anyone can call this action. 
