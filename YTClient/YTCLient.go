package YTClient

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/0xFA7E/foxeye/SqliteClient"

	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"
)

//a const for building video links
const watchyt = "https://www.youtube.com/watch?v="

/*YoutubeClient is filled with an api key, then call Service() to generate a youtube api
client. Methods can then be attached to call the api such as RecentVideo */
type YoutubeClient struct {
	APIKey  string
	service *youtube.Service
	SQLCli  *SqliteClient.SQLCli
}

func (c *YoutubeClient) Service() {
	client := &http.Client{Transport: &transport.APIKey{Key: c.APIKey}}
	ser, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating youtube client: %v", err)
	}
	c.service = ser
}

func (c *YoutubeClient) ExtractID(url string) (string, error) {

	s := strings.Split(url, "/")
	if strings.Contains(url, "/channel/") {
		c.SQLCli.UpdateChanID(url, s[len(s)-1])
		return s[len(s)-1], nil
	}
	call := c.service.Channels.List("id")
	call = call.ForUsername(s[len(s)-1])
	resp, err := call.Do()
	if err != nil {
		return "", err
	}
	if len(resp.Items) <= 0 {
		return "", errors.New("Not a valid channel link")
	}
	c.SQLCli.UpdateChanID(url, resp.Items[0].Id)
	return resp.Items[0].Id, nil
}

func (c *YoutubeClient) RecentVideo(timeAfter time.Time) []string {
	//fmt.Printf("Checking results of %v at %v\n", channel.channelID, YTime(time.Now()))
	//building our request, first is type of info we want
	var urls []string
	channels := c.SQLCli.WatchList()
	channelList := strings.Join(channels[:], ",")
	call := c.service.Channels.List("contentDetails")
	//we want grouped results of channel IDs hopefully conserves quota
	call = call.Id(channelList)
	response, err := call.Do()
	if err != nil {
		log.Printf("Failure executing request: %v", err)
		return []string{}
	}
	if len(response.Items) == 0 {
		return []string{}
	}
	for _, respItem := range response.Items {
		respChan := respItem.Id
		uploadsID := respItem.ContentDetails.RelatedPlaylists.Uploads

		playlistCall := c.service.PlaylistItems.List("contentDetails")
		playlistCall = playlistCall.PlaylistId(uploadsID)
		playlistResponse, err := playlistCall.Do()
		if err != nil {
			log.Printf("\n\n[!]Failure executing request: %v\n\n", err)
			return []string{}
		}
		if len(playlistResponse.Items) == 0 {
			return []string{}
		}
		publishedAt := playlistResponse.Items[0].ContentDetails.VideoPublishedAt
		rVid := playlistResponse.Items[0].ContentDetails.VideoId
		isNew := c.SQLCli.UpdateVideo(respChan, rVid, publishedAt)
		if isNew == true {
			urls = append(urls, createLink(rVid))
		}
	}
	return urls
}

func (c *YoutubeClient) AddChannels(channels []string) error {
	for _, channel := range channels {
		channelID, err := c.ExtractID(channel)
		if err != nil {
			return errors.New("Failed to extract channel id")
		}
		err = c.SQLCli.AddChannel(channel, channelID)
		if err != nil {
			return errors.New("Failed to add channel")
		}
	}
	return nil
}

func createLink(vID string) string {
	url := watchyt + vID
	return url
}

/*yTime returns a youtube approved time string from a time.Time format*/
func YTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)

}
