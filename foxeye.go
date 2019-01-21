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

func setupClients(c *config) DiscordClient.DiscordClient {
	ytClient := YTClient.Service(c.YtAPIKey)
	sqldb := SqliteClient.InitDB(c.Database)
	dclient := DiscordClient.DiscordClient{APIKey: c.DiscordAPIKey, PostChannel: c.PostChannel, Log: log, WatchClient: ytClient, DatabaseClient: sqldb}
	err := dclient.Init()
	if err != nil {
		log.WithFields(logrus.Fields{
			"ytClient": ytClient,
			"sqldb":    sqldb,
			"dclient":  dclient,
			"error":    err,
		}).Fatal("Could not intitialize Clients")
	}
	return dclient
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

	/*
		fmt.Printf("\nChannels to Watch:")
		chanlist, _ := reader.ReadString('\n')
		dclient := setupClients(&configRaw)
		//ytClient.(strings.Split(strings.TrimSpace(chanlist), " "))
	*/
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

	dClient := setupClients(&configuration)
	defer dClient.Close()
	log.WithFields(logrus.Fields{
		"Discord Client": dClient,
		"Youtube Client": dClient.WatchClient,
		"SQL Client":     dClient.DatabaseClient,
	}).Debug("Bot is now running")

	go dClient.MonitorUploads(ytTimeRate)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Info("Closing bot")
	dClient.SendByName(configuration.PostChannel, "NOOO I DONT WANT TO DI----")
}
