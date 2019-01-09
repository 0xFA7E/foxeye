package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/0xFA7E/foxeye/DiscordClient"
	"github.com/0xFA7E/foxeye/SqliteClient"
	"github.com/0xFA7E/foxeye/YTClient"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type config struct {
	/*struct for loading and writing config info*/
	YtAPIKey      string `json:"ytAPIKey"`
	DiscordAPIKey string `json:"discordAPIKey"`
	PostChannel   string `json:"postchannel"`
	Database      string `json:"database"`
}

func (s config) IsEmpty() bool {
	return reflect.DeepEqual(s, config{})
}

func setupClients(ytKey string, dkey string, filename string) (YTClient.YoutubeClient, DiscordClient.ServerContext, SqliteClient.ChannelWatch) {
	ytClient := YTClient.YoutubeClient{APIKey: ytKey}
	ytClient.Service()
	dclient := DiscordClient.ServerContext{APIKey: dkey}
	dclient.Init()
	sqldb := SqliteClient.ChannelWatch{}
	sqldb.InitDB(filename)
	sqldb.CreateTable()
	ytClient.SQLCli = &sqldb

	return ytClient, dclient, sqldb
}

func monitorUploads(configuration config, ytClient YTClient.YoutubeClient, dClient DiscordClient.ServerContext, rate time.Duration) {
	for {
		t := time.Now()
		timer := time.NewTimer(rate * time.Second)
		<-timer.C
		videos := ytClient.RecentVideo(t)
		if videos != nil {
			for _, v := range videos {
				err := dClient.SendByName(configuration.PostChannel, v)
				if err != nil {
					log.WithFields(logrus.Fields{
						"Post Channel": configuration.PostChannel,
						"Video":        v,
					}).Error("Error sending message")
				}
			}
		}
	}
}

func genConfig(filename string) {
	configRaw := config{}
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Youtube API Key:")
	ytkey, _ := reader.ReadString('\n')
	configRaw.YtAPIKey = strings.TrimSpace(ytkey)

	fmt.Printf("\nDiscord API Key:")
	dkey, _ := reader.ReadString('\n')
	configRaw.DiscordAPIKey = strings.TrimSpace(dkey)

	fmt.Printf("\nChannel to post in:")
	postchan, _ := reader.ReadString('\n')
	configRaw.PostChannel = strings.TrimSpace(postchan)

	fmt.Printf("\nsqlite DB file to use:")
	dbfile, _ := reader.ReadString('\n')
	configRaw.Database = strings.TrimSpace(dbfile)
	db := SqliteClient.ChannelWatch{}
	db.InitDB(configRaw.Database)
	db.CreateTable()

	fmt.Printf("\nChannels to Watch:")
	chanlist, _ := reader.ReadString('\n')
	db.AddChannels(strings.Split(strings.TrimSpace(chanlist), " "))

	configFile, err := os.Create(filename)
	if err != nil {
		log.WithFields(logrus.Fields{
			"filename:": filename,
			"reason":    err,
		}).Fatal("Could not create config")
	}
	defer configFile.Close()
	configJSON, _ := json.MarshalIndent(configRaw, "", "    ")
	_, werr := configFile.Write(configJSON)
	if werr != nil {
		log.WithFields(logrus.Fields{
			"filename": filename,
			"reason":   werr,
			"json":     configJSON,
		}).Fatal("Could not write json data to config file")
	}
	configFile.Sync()
	fmt.Println("Configuration written to disk")

	os.Exit(0)
}

func main() {
	ytTimeRate := time.Duration(5) //only change if you know what youre doing, greatly affects quota!

	configuration := config{}
	configFileName := flag.String("config", "", "Configuration file")
	generateConfig := flag.String("gen", "", "Generate a config file with output name")
	flag.Parse()
	if *generateConfig != "" {
		genConfig(*generateConfig)

	} else if *configFileName != "" {
		configFile, oerr := os.Open(*configFileName)
		if oerr != nil {
			log.WithFields(logrus.Fields{
				"filename": *configFileName,
				"reason":   oerr,
			}).Fatal("Failed to open config file")
		}
		defer configFile.Close()
		byteValue, rerr := ioutil.ReadAll(configFile)
		if rerr != nil {
			log.WithFields(logrus.Fields{
				"reason": rerr,
			}).Fatal("Failed to read config file")
		}
		json.Unmarshal(byteValue, &configuration)
		if configuration.IsEmpty() {
			log.Fatal("Failed to load configuration file. Likely not formatted correctly")
		}
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
	log.Info("FOXEYE ENGAGE")

	ytClient, dClient, sqlClient := setupClients(configuration.YtAPIKey, configuration.DiscordAPIKey, configuration.Database)

	log.WithFields(logrus.Fields{
		"Youtube Client": ytClient,
		"Discord Client": dClient,
		"SQL Client":     sqlClient,
	}).Debug("Bot is now running")

	//pop out an event once guildCreate has run
	err := dClient.SendByName(configuration.PostChannel, "Foxeye is on the watch!")
	if err != nil {
		log.WithFields(logrus.Fields{
			"reason": err,
		}).Error("Failed sending test message")
	}
	go monitorUploads(configuration, ytClient, dClient, ytTimeRate)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Info("Closing bot")
	dClient.SendByName("test", "NOOO I DONT WANT TO DI----")
}
