package feeds

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
)

type WebsiteFeed struct {
	feedURL string
	limit   int
	items   []DashboardItem
}

func NewWebsiteFeed(feedURL string, limit int) *WebsiteFeed {
	return &WebsiteFeed{
		feedURL: feedURL,
		limit:   limit,
		items:   make([]DashboardItem, 0),
	}
}

func (w *WebsiteFeed) ScheduleRefresh(ctx context.Context) *cron.Cron {
	return ScheduleRefresh(ctx, w)
}

func (w *WebsiteFeed) TransformFeedItems(ctx context.Context, feed *gofeed.Feed) {
	TransformFeedItems(w, feed)
}

func (w *WebsiteFeed) SetItems(items []DashboardItem) { w.items = items }
func (w *WebsiteFeed) PullItems(ctx context.Context)  { PullItems(ctx, w) }
func (w *WebsiteFeed) PrintItems()                    { PrintItems(w) }
func (w *WebsiteFeed) Items() []DashboardItem         { return w.items }
func (w *WebsiteFeed) FeedURL() string                { return w.feedURL }
func (w *WebsiteFeed) Limit() int                     { return w.limit }
