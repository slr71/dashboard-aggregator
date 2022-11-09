package feeds

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type InstantLaunchesFeed struct {
	WebsiteFeed
	Path           string
	AppExposerUser string
	Attribute      string
	Value          string
}

func NewInstantLaunchesFeed(feedURL string, limit int, user string) *InstantLaunchesFeed {
	return &InstantLaunchesFeed{
		WebsiteFeed: WebsiteFeed{
			FeedURL: feedURL,
			Limit:   limit,
		},
		Path:           "/instantlaunches/metadata/full",
		Attribute:      "ui_location",
		Value:          "dashboard",
		AppExposerUser: user,
	}
}

func (i *InstantLaunchesFeed) PullItems(ctx context.Context) {
	u, err := url.Parse(i.WebsiteFeed.FeedURL)
	if err != nil {
		log.Error(err)
		return

	}
	u.Path = i.Path

	q := u.Query()
	q.Set("user", i.AppExposerUser)
	q.Set("attribute", i.Attribute)
	q.Set("value", i.Value)
	u.RawQuery = q.Encode()

	log.Infof("pulling items from %s", u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		log.Error(err)
		resp.Body.Close()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return
		}
		log.Error(string(msg))
		return
	}

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return
	}

	if err = json.Unmarshal(msg, &i.WebsiteFeed.Items); err != nil {
		log.Error(err)
		return
	}
}

func (i *InstantLaunchesFeed) GetItems(ctx context.Context) []DashboardItem {
	i.PullItems(ctx)
	return i.WebsiteFeed.Items
}
