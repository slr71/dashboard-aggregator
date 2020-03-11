/**
 * @author johnworth
 *
 * Pulls in the news feed and makes it available to the dashboard.
 *
 * @module feeds
 */
import path from "path";
import Parser from "rss-parser";

const transformFeedItem = (item) => {
    return {
        id: item.guid,
        name: item.title,
        description: item.contentSnippet,
        link: item.link,
        dateAdded: item.isoDate,
    };
};

export const feedURL = (baseURL, feedPath) => {
    const u = new URL(baseURL);
    u.pathname = path.join(u.pathname, feedPath);
    return u.toString();
};

export default class WebsiteFeed {
    constructor(feedURL, limit = 20) {
        this.feedURL = feedURL;
        this.limit = limit;
        this.items = [];
        this.pullItems();
    }

    async pullItems() {
        const parser = new Parser();
        const feed = await parser.parseURL(this.feedURL);

        // Make sure the latest post is first.
        feed.items.reverse();

        if (feed.items.length > this.limit) {
            for (let i = 0; i < this.limit; i++) {
                this.items.push(transformFeedItem(feed.items[i]));
            }
        } else {
            this.items = [...feed.items.map((item) => transformFeedItem(item))];
        }
    }

    async getItems() {
        if (this.items.length === 0) {
            await this.pullItems();
        }
        return this.items;
    }
}
