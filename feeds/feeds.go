package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cyverse-de/go-mod/logging"
	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
	"github.com/samber/lo"
)

var log = logging.Log.WithField("package", "feeds")

type DashboardItem struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	DateAdded       string `json:"date_added"`
	Author          string `json:"author"`
	PublicationDate string `json:"publication_date"`
	Content         string `json:"content"`
	Link            string `json:"link"`
	ThumbnailURL    string `json:"thumbnailUrl"`
}

type DashboardFeeder interface {
	FeedURL() string
	Items() []DashboardItem
	SetItems(items []DashboardItem)
	Limit() int
	PrintItems()

	ScheduleRefresh(ctx context.Context) (*cron.Cron, error)
	PullItems(ctx context.Context)
	TransformFeedItems(ctx context.Context, feed *gofeed.Feed) []DashboardItem
}

const InstantLaunchesFeedName = "instant-launches"
const NewsFeedName = "news"
const EventsFeedName = "events"
const VideosFeedName = "videos"

// PublicFeeds manages a set of feeds from external sources. While there is some
// overlap with the DashboardFeeder interface, it does not fully implement it.
type PublicFeeds struct {
	feeders map[string]DashboardFeeder
	crons   []*cron.Cron
}

func NewPublicFeeds() *PublicFeeds {
	return &PublicFeeds{
		feeders: make(map[string]DashboardFeeder),
		crons:   make([]*cron.Cron, 0),
	}
}

func (p PublicFeeds) AddFeed(ctx context.Context, name string, feeder DashboardFeeder) {
	p.feeders[name] = feeder
}

func (p PublicFeeds) Names() []string {
	return lo.Keys(p.feeders)
}

func (p PublicFeeds) PullItems(ctx context.Context) {
	for _, feeder := range p.feeders {
		feeder.PullItems(ctx)
	}
}

func (p PublicFeeds) PrintItems() {
	for _, feeder := range p.feeders {
		feeder.PrintItems()
	}
}

func (p PublicFeeds) ScheduleRefreshes(ctx context.Context) error {
	for _, feeder := range p.feeders {
		c, err := feeder.ScheduleRefresh(ctx)
		if err != nil {
			return err
		}
		p.crons = append(p.crons, c) // nolint:all
	}
	return nil
}

func (p PublicFeeds) Items(ctx context.Context, name string) []DashboardItem {
	return p.feeders[name].Items()
}

func (p PublicFeeds) Marshallable(ctx context.Context) map[string][]DashboardItem {
	retval := make(map[string][]DashboardItem)

	for name, feeder := range p.feeders {
		retval[name] = feeder.Items()
	}

	return retval
}

/*
 *
 * Utility functions for other implementations of DashboardFeeder to use.
 *
 */

func ScheduleRefresh(ctx context.Context, f DashboardFeeder) (*cron.Cron, error) {
	log := log.WithField("context", "scheduling feed refresh")

	j := cron.New()

	log.Infof("scheduling a refresh of items from %s", f.FeedURL())

	_, err := j.AddFunc("0 * * * *", func() {
		log.Infof("starting refresh of %s", f.FeedURL())
		PullItems(ctx, f)
	})
	if err != nil {
		return nil, err
	}

	j.Start()

	log.Debugf("done scheduling a refresh of items from %s", f.FeedURL())

	return j, nil
}

func TransformFeedItems(f DashboardFeeder, feed *gofeed.Feed) []DashboardItem {
	log.Infof("transforming feed items from %s", f.FeedURL())

	items := lo.Map(feed.Items, func(in *gofeed.Item, index int) DashboardItem {
		// log.Debugf("content %s", in.Content)
		// log.Debugf("description %s", in.Description)

		// descLength := 281
		// if len(in.Content) <= descLength {
		// 	descLength = len(in.Content)
		// }

		description := fmt.Sprintf("%s\n%s\n%s", in.Title, in.Author.Name, in.PublishedParsed.Format(time.RFC1123))
		dbi := DashboardItem{
			ID:              in.GUID,
			Name:            in.Title,
			Description:     description,
			DateAdded:       in.PublishedParsed.Format(time.RFC3339),
			Author:          in.Author.Name,
			PublicationDate: in.Published,
			Content:         in.Description,
			Link:            in.Link,
		}
		return dbi
	})

	log.Debugf("done transforming feed items from %s", f.FeedURL())

	return items
}

func PrintItems(f DashboardFeeder) {
	log := log.WithField("context", "printing items")

	for _, item := range f.Items() {
		b, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			log.Error(err)
			return
		}
		log.Info(string(b))
	}

	log.Infof("done printing items from %s", f.FeedURL())
}

func PullItems(ctx context.Context, f DashboardFeeder) {
	log := log.WithField("context", "pulling items")

	log.Infof("pulling feed items from %s", f.FeedURL())

	p := gofeed.NewParser()

	feed, err := p.ParseURLWithContext(f.FeedURL(), ctx)
	if err != nil {
		log.Error(err)
		return
	}

	//feed.Items = lo.Reverse(feed.Items)

	if len(feed.Items) > f.Limit() {
		feed.Items = feed.Items[0 : f.Limit()+1]
	}

	items := f.TransformFeedItems(ctx, feed)
	f.SetItems(items)
}
