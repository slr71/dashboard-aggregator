package feeds

import (
	"context"
	"sync"

	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
)

type WebsiteFeed struct {
	feedURL string
	limit   int
	items   []DashboardItem
	mu      sync.RWMutex
}

func NewWebsiteFeed(feedURL string, limit int) *WebsiteFeed {
	return &WebsiteFeed{
		feedURL: feedURL,
		limit:   limit,
		items:   make([]DashboardItem, 0),
		mu:      sync.RWMutex{},
	}
}

func (w *WebsiteFeed) ScheduleRefresh(ctx context.Context) *cron.Cron {
	return ScheduleRefresh(ctx, w)
}

func (w *WebsiteFeed) TransformFeedItems(ctx context.Context, feed *gofeed.Feed) []DashboardItem {
	return TransformFeedItems(w, feed)
}

func (w *WebsiteFeed) SetItems(items []DashboardItem) {
	w.mu.Lock()
	w.items = items
	w.mu.Unlock()
}
func (w *WebsiteFeed) PullItems(ctx context.Context) {
	PullItems(ctx, w)

}
func (w *WebsiteFeed) PrintItems() {
	PrintItems(w)
}
func (w *WebsiteFeed) Items() []DashboardItem {
	w.mu.RLock()
	retval := w.items
	w.mu.RUnlock()
	return retval
}
func (w *WebsiteFeed) FeedURL() string { return w.feedURL }
func (w *WebsiteFeed) Limit() int      { return w.limit }
