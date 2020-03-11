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
    }

    async pullItems() {
        console.log(this.feedURL);
        const parser = new Parser();
        const feed = await parser.parseURL(this.feedURL);

        if (feed.items.length > this.limit) {
            console.log("using for loop");
            for (let i = 0; i < this.limit; i++) {
                this.items.push(transformFeedItem(feed.items[i]));
            }
        } else {
            console.log("using spread");
            this.items = [...feed.items.map((item) => transformFeedItem(item))];
        }
    }

    async getItems() {
        if (this.items.length === 0) {
            console.log("calling pullItems");
            await this.pullItems();
        }
        return this.items;
    }
}
