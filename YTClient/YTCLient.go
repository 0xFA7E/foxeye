package YTClient

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/0xFA7E/foxeye/DiscordClient"
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
	//SQLCli  *SqliteClient.SQLCli
}

func Service(apikey string) *YoutubeClient {
	client := &http.Client{Transport: &transport.APIKey{Key: apikey}}
	ser, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating youtube client: %v", err)
	}
	ytClient := YoutubeClient{APIKey: apikey, service: ser}
	return &ytClient
}

func (c *YoutubeClient) RecentVideo(channelIds []string, timeAfter time.Time) ([]DiscordClient.Video, error) {
	//fmt.Printf("Checking results of %v at %v\n", channel.channelID, YTime(time.Now()))
	//building our request, first is type of info we want
	var videos []DiscordClient.Video
	channelList := strings.Join(channelIds[:], ",")
	call := c.service.Channels.List("contentDetails")
	//we want grouped results of channel IDs hopefully conserves quota
	call = call.Id(channelList)
	response, err := call.Do()
	if err != nil {
		log.Printf("Failure executing request: %v", err)
		return []DiscordClient.Video{}, err
	}
	if len(response.Items) == 0 {
		return []DiscordClient.Video{}, err
	}
	for _, respItem := range response.Items {
		respChan := respItem.Id
		uploadsID := respItem.ContentDetails.RelatedPlaylists.Uploads

		playlistCall := c.service.PlaylistItems.List("contentDetails")
		playlistCall = playlistCall.PlaylistId(uploadsID)
		playlistResponse, err := playlistCall.Do()
		if err != nil {
			log.Printf("\n\n[!]Failure executing request: %v\n\n", err)
			return []DiscordClient.Video{}, err
		}
		if len(playlistResponse.Items) == 0 {
			return []DiscordClient.Video{}, nil
		}
		publishedAt := playlistResponse.Items[0].ContentDetails.VideoPublishedAt
		rVid := playlistResponse.Items[0].ContentDetails.VideoId
		//isNew := c.SQLCli.UpdateVideo(respChan, rVid, publishedAt)
		link := CreateLink(rVid)
		v := &Video{}
		v.New(respChan, link, publishedAt)
		videos = append(videos, v)
	}
	return videos, nil
}

func (c *YoutubeClient) ExtractIDs(urls []string) (map[string]string, error) {
	ids := make(map[string]string)
	for _, v := range urls {
		s := strings.Split(v, "/")
		if strings.Contains(v, "/channel/") {
			//c.SQLCli.UpdateChanID(v, s[len(s)-1])
			ids[v] = s[len(s)-1]
		}
		call := c.service.Channels.List("id")
		call = call.ForUsername(s[len(s)-1])
		resp, err := call.Do()
		if err != nil {
			//TODO add logging
			continue
		}
		if len(resp.Items) <= 0 {
			//TODO add logging
			//return "", errors.New("Not a valid channel link")
			continue
		}
		//c.SQLCli.UpdateChanID(url, resp.Items[0].Id)
		ids[v] = resp.Items[0].Id
	}
	return ids, nil
}

/*
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
*/
func CreateLink(vID string) string {
	url := watchyt + vID
	return url
}

/*yTime returns a youtube approved time string from a time.Time format*/
func YTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)

}
