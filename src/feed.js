/**
 * @author johnworth
 *
 * Pulls in the news feed and makes it available to the dashboard.
 *
 * @module feeds
 */
import path from "path";
import Parser from "rss-parser";
import { CronJob } from "cron";
import logger from "./logging";

const transformFeedItem = ({
    guid: id,
    title: name,
    contentSnippet: description,
    isoDate: date_added,
    creator,
    pubDate: publication_date,
    content,
}) => ({
    id,
    name,
    description,
    date_added,
    creator,
    publication_date,
    content,
});

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

    scheduleRefresh() {
        const job = new CronJob("0 * * * *", () => {
            logger.info(`starting refresh of ${this.feedURL}`);
            this.pullItems();
        });
        return job;
    }

    async pullItems() {
        logger.info(`pulling items from ${this.feedURL}`);

        const parser = new Parser();
        const feed = await parser.parseURL(this.feedURL);

        // Make sure the latest post is first.
        feed.items.reverse();

        if (feed.items.length > this.limit) {
            logger.debug(`using for-loop population for ${this.feedURL}`);

            let newList = [];
            for (let i = 0; i < this.limit; i++) {
                newList.push(transformFeedItem(feed.items[i]));
            }
            this.items = [...newList]; // spread isn't technically necessary, but I prefer it to reference copying.
        } else {
            logger.debug(`using map-spread population for ${this.feedURL}`);
            this.items = [...feed.items.map((item) => transformFeedItem(item))];
        }

        logger.info(`done pulling items from ${this.feedURL}`);
    }

    async getItems() {
        if (this.items.length === 0) {
            await this.pullItems();
        }
        return this.items;
    }
}
