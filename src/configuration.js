/**
 * Configuration for the dashboard-aggregator.
 *
 * @module configuration
 */
import config from "config";

/**
 * Verifies that a setting is present in the configuration.
 *
 * @param {string} name
 */
function validateConfigSetting(name) {
    if (!config.has(name) || config.get(name) === null) {
        throw Error(`${name} must be set in the configuration`);
    }
}

/**
 * Validates that required settings are present in the configuration.
 */
const validate = () => {
    validateConfigSetting("db.user");
    validateConfigSetting("db.password");
    validateConfigSetting("db.host");
    validateConfigSetting("db.port");
    validateConfigSetting("db.database");
    validateConfigSetting("listen_port");
    validateConfigSetting("permissions.uri");
    validateConfigSetting("permissions.public_group");
};

validate();

/**
 * The database user.
 *
 * @type {string}
 */
export const dbUser = config.get("db.user");

/**
 * The database password.
 *
 * @type {string}
 */
export const dbPass = config.get("db.password");

/**
 * The database host.
 *
 * @type {string}
 */
export const dbHost = config.get("db.host");

/**
 * The database port.
 *
 * @type {string}
 */
export const dbPort = parseInt(config.get("db.port"), 10);

/**
 * The database name.
 *
 * @type {string}
 */
export const dbDatabase = config.get("db.database");

/**
 * The logging level.
 *
 * @type {string}
 */
export const logLevel = config.get("logging.level") || "info";

/**
 * The label to use for logging.
 */
export const logLabel = config.get("logging.label") || "dashboard-aggregator";

/**
 * The listen port for the app.
 */
export const listenPort = parseInt(config.get("listen_port"), 10);

/**
 * The base URI for the permissions service.
 */
export const permissionsURL = config.get("permissions.uri");

/**
 * The public group in the permissions service.
 */
export const publicGroup = config.get("permissions.public_group");

/**
 * The base URL for the website.
 */
export const websiteURL = config.get("website.url");

/**
 * The path to the news RSS feed file on the website.
 */
export const newsFeedPath = config.get("website.feeds.news");

/**
 * The path to the events RSS feed file on the website.
 */
export const eventsFeedPath = config.get("website.feeds.events");

/**
 * The URL to the RSS feed for the videos.
 */
export const videosURL = config.get("videos.url");
