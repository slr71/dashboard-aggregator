package feeds

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/samber/lo"
)

type VideoFeed struct {
	WebsiteFeed
}

func NewVideoFeed(feedURL string, limit int) *VideoFeed {
	return &VideoFeed{
		WebsiteFeed: WebsiteFeed{
			FeedURL: feedURL,
			Limit:   limit,
			Items:   make([]DashboardItem, 0),
		},
	}
}

func (v *VideoFeed) TransformFeedItems(ctx context.Context, feed *gofeed.Feed) {
	v.WebsiteFeed.Items = lo.Map(feed.Items, func(in *gofeed.Item, index int) DashboardItem {
		dbi := DashboardItem{
			ID:              in.GUID,
			Name:            in.Title,
			Description:     in.Extensions["media:group"]["media:description"][0].Value,
			DateAdded:       in.Published,
			Author:          in.Author.Name,
			PublicationDate: in.Published,
			Content:         in.Content,
			Link:            in.Link,
			ThumbnailURL:    in.Extensions["media:group"]["media:thumbnail"][0].Attrs["$"],
		}
		log.Debugf("%+v", in)
		return dbi
	})
}
