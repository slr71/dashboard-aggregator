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

const transformFeedItem = (item) => {
    const {
        guid: id,
        title: name,
        contentSnippet: description,
        isoDate: date_added,
        author,
        pubDate: publication_date,
        content,
        link,
    } = item;

    return {
        id,
        name,
        description,
        date_added,
        author,
        publication_date,
        content,
        link,
    };
};

const transformVideoItem = (item) => {
    const {
        id,
        title: name,
        isoDate: date_added,
        author,
        pubDate: publication_date,
        link,
    } = item;
    const description = item["media:group"]["media:description"][0];
    const thumbnailUrl = item["media:group"]["media:thumbnail"][0]["$"].url;

    return {
        id,
        name,
        description,
        date_added,
        author,
        publication_date,
        link,
        thumbnailUrl,
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

    scheduleRefresh() {
        const job = new CronJob("0 * * * *", () => {
            logger.info(`starting refresh of ${this.feedURL}`);
            this.pullItems();
        });
        return job;
    }

    async pullItems() {
        logger.info(`pulling items from ${this.feedURL}`);

        const parser = new Parser({
            customFields: {
                item: [
                    ["dc:creator", "author"],
                    ["description", "content", { includeSnippet: true }],
                ],
            },
        });
        const feed = await parser.parseURL(this.feedURL);

        // Make sure the latest post is first.
        feed.items.reverse();

        if (feed.items.length > this.limit) {
            logger.debug(`using for-loop population for ${this.feedURL}`);

            let newList = [];
            for (let i = 0; i < this.limit; i++) {
                newList.push(transformFeedItem(feed.items[i]));
            }
            this.items = [...newList];
        } else {
            logger.debug(`using map-spread population for ${this.feedURL}`);
            this.items = [...feed.items.map((item) => transformFeedItem(item))];
        }

        logger.info(`done pulling items from ${this.feedURL}`);
    }

    // Useful for debugging.
    async printItems() {
        logger.info(`printing items from ${this.feedURL}`);

        const parser = new Parser();
        const feed = await parser.parseURL(this.feedURL);

        feed.items.reverse();

        feed.items.forEach((item) => {
            console.log("\n");
            console.log(JSON.stringify(item, null, 2));
        });

        logger.info(`done printing items from ${this.feedURL}`);
    }

    async getItems() {
        if (this.items.length === 0) {
            await this.pullItems();
        }
        return this.items;
    }
}

export class VideoFeed extends WebsiteFeed {
    constructor(feedURL, limit = 20) {
        super(feedURL, limit);
    }

    async pullItems() {
        logger.info(`pulling items from ${this.feedURL}`);

        const parser = new Parser({
            customFields: {
                item: [["media:group", "media:group"]],
            },
        });
        const feed = await parser.parseURL(this.feedURL);

        // Make sure the latest post is first.
        feed.items.reverse();

        if (feed.items.length > this.limit) {
            logger.debug(`using for-loop population for ${this.feedURL}`);

            let newList = [];
            for (let i = 0; i < this.limit; i++) {
                newList.push(transformVideoItem(feed.items[i]));
            }
            this.items = [...newList];
        } else {
            logger.debug(`using map-spread population for ${this.feedURL}`);
            this.items = [
                ...feed.items.map((item) => transformVideoItem(item)),
            ];
        }

        logger.info(`done pulling items from ${this.feedURL}`);
    }

    // Useful for debugging.
    async printItems() {
        logger.info(`printing items from ${this.feedURL}`);

        const parser = new Parser({
            customFields: {
                item: [["media:group", "media:group"]],
            },
        });
        const feed = await parser.parseURL(this.feedURL);

        feed.items.reverse();

        feed.items.forEach((item) => {
            console.log("\n");
            console.log(JSON.stringify(transformVideoItem(item), null, 2));
        });

        logger.info(`done printing items from ${this.feedURL}`);
    }
}
