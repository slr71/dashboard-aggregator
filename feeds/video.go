package feeds

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/robfig/cron/v3"
	"github.com/samber/lo"
)

type VideoFeed struct {
	feedURL string
	limit   int
	items   []DashboardItem
}

func NewVideoFeed(feedURL string, limit int) *VideoFeed {
	return &VideoFeed{
		feedURL: feedURL,
		limit:   limit,
		items:   make([]DashboardItem, 0),
	}
}

func (v *VideoFeed) Items() []DashboardItem         { return v.items }
func (v *VideoFeed) SetItems(items []DashboardItem) { v.items = items }
func (v *VideoFeed) Limit() int                     { return v.limit }
func (v *VideoFeed) FeedURL() string                { return v.feedURL }
func (v *VideoFeed) PrintItems()                    { PrintItems(v) }
func (v *VideoFeed) PullItems(ctx context.Context)  { PullItems(ctx, v) }

func (v *VideoFeed) ScheduleRefresh(ctx context.Context) *cron.Cron {
	return ScheduleRefresh(ctx, v)
}

func (v *VideoFeed) TransformFeedItems(ctx context.Context, feed *gofeed.Feed) {

	log.Info("transforming video feed items")

	v.SetItems(lo.Map(feed.Items, func(in *gofeed.Item, index int) DashboardItem {
		var (
			description  string
			thumbnailURL string
		)

		if media, ok := in.Extensions["media"]; ok {
			if groups, ok := media["group"]; ok {
				if len(groups) > 0 {
					group := groups[0]
					if descs, ok := group.Children["description"]; ok {
						if len(descs) > 0 {
							desc := descs[0]
							description = desc.Value
						}
					}

					if thumbs, ok := group.Children["thumbnail"]; ok {
						if len(thumbs) > 0 {
							thumb := thumbs[0]
							thumbnailURL = thumb.Value
						}
					}
				}
			}
		}

		dbi := DashboardItem{
			ID:              in.GUID,
			Name:            in.Title,
			Description:     description,
			DateAdded:       in.Published,
			Author:          in.Author.Name,
			PublicationDate: in.Published,
			Content:         in.Content,
			Link:            in.Link,
			ThumbnailURL:    thumbnailURL,
		}

		return dbi
	}))

	log.Debug("done transforming video feed items")
}
