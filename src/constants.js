/**
 * Constants for the dashboard-aggregator service.
 *
 * @module constants
 */

import * as config from "./configuration";

export default {
    DEFAULT_START_DATE_INTERVAL: "1 year",
    FEATURED_APPS_AVUS: [
        {
            attr: config.featuredAppsAttr,
            value: config.featuredAppsValue,
        },
    ],
};
