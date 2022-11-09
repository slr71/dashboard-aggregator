package feeds

import (
	"context"
	"encoding/json"

	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
	"github.com/samber/lo"
)

type WebsiteFeed struct {
	FeedURL string
	Limit   int
	Items   []DashboardItem
}

func NewWebsiteFeed(feedURL string, limit int) *WebsiteFeed {
	return &WebsiteFeed{
		FeedURL: feedURL,
		Limit:   limit,
		Items:   make([]DashboardItem, 0),
	}
}

func (w *WebsiteFeed) ScheduleRefresh(ctx context.Context) *cron.Cron {
	log := log.WithField("context", "scheduling feed refresh")
	j := cron.New()
	j.AddFunc("0 * * * *", func() {
		log.Infof("starting refresh of %s", w.FeedURL)
		w.PullItems(ctx)
	})
	j.Start()
	return j
}

func (w *WebsiteFeed) TransformFeedItems(ctx context.Context, feed *gofeed.Feed) {
	w.Items = lo.Map(feed.Items, func(in *gofeed.Item, index int) DashboardItem {
		descLength := 281
		if len(in.Content) <= descLength {
			descLength = len(in.Content)
		}
		dbi := DashboardItem{
			ID:              in.GUID,
			Name:            in.Title,
			Description:     in.Content[0:descLength],
			DateAdded:       in.Published,
			Author:          in.Author.Name,
			PublicationDate: in.Published,
			Content:         in.Content,
			Link:            in.Link,
		}
		return dbi
	})

}

func (w *WebsiteFeed) PullItems(ctx context.Context) {
	log := log.WithField("context", "pulling items")
	p := gofeed.NewParser()
	feed, err := p.ParseURLWithContext(w.FeedURL, ctx)
	if err != nil {
		log.Error(err)
		return
	}

	feed.Items = lo.Reverse(feed.Items)

	if len(feed.Items) > w.Limit {
		feed.Items = feed.Items[0 : w.Limit+1]
	}

	w.TransformFeedItems(ctx, feed)
}

func (w *WebsiteFeed) PrintItems(ctx context.Context) {
	log := log.WithField("context", "printing items")

	p := gofeed.NewParser()
	feed, err := p.ParseURLWithContext(w.FeedURL, ctx)
	if err != nil {
		log.Error(err)
		return
	}

	feed.Items = lo.Reverse(feed.Items)

	w.TransformFeedItems(ctx, feed)

	for _, item := range feed.Items {
		b, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			log.Error(err)
			return
		}
		log.Info(string(b))
	}

	log.Infof("done printing items from %s", w.FeedURL)
}

func (w *WebsiteFeed) GetItems(ctx context.Context) []DashboardItem {
	if len(w.Items) == 0 {
		w.PullItems(ctx)
	}
	return w.Items
}
