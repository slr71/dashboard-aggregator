package feeds

import (
	"context"

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
	ThumbnailURL    string `json:"thumbnail_url"`
}

type DashboardFeeder interface {
	ScheduleRefresh(ctx context.Context) *cron.Cron
	PullItems(ctx context.Context)
	PrintItems(ctx context.Context)
	GetItems(ctx context.Context) []DashboardItem
	TransformFeedItems(ctx context.Context, feed *gofeed.Feed)
}

const InstantLaunchesFeedName = "instant-launches"
const NewsFeedName = "news"
const EventsFeedName = "events"
const VideosFeedName = "videos"

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

func (p PublicFeeds) ScheduleRefreshes(ctx context.Context) {
	for _, feeder := range p.feeders {
		c := feeder.ScheduleRefresh(ctx)
		p.crons = append(p.crons, c)
	}
}

func (p PublicFeeds) FeedItems(ctx context.Context, name string) []DashboardItem {
	return p.feeders[name].GetItems(ctx)
}

func (p PublicFeeds) Marshallable(ctx context.Context) map[string][]DashboardItem {
	retval := make(map[string][]DashboardItem)

	for name, feeder := range p.feeders {
		retval[name] = feeder.GetItems(ctx)
	}

	return retval
}
