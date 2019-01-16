package YTClient

type Video struct {
	channelID   string
	videoLink   string
	publishedAt string
}

func (v *Video) ChannelID() string {
	return v.channelID
}

func (v *Video) VideoLink() string {
	return v.videoLink
}

func (v *Video) PublishTime() string {
	return v.publishedAt
}

func (v *Video) New(channel string, video string, published string) {
	v.channelID = channel
	v.videoLink = video
	v.publishedAt = published
}
