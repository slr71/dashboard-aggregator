import express from "express";
import { Client } from "pg";

import * as config from "./configuration";
import logger, { errorLogger, requestLogger } from "./logging";
import recentlyAddedHandler from "./apps/recentlyAdded";
import publicAppsHandler from "./apps/public";
import recentAnalysesHandler from "./analyses/recent";
import runningAnalysesHandler from "./analyses/running";

logger.info("creating database client");

// Set up the database connection. May have to change to a Pool in the near future.
const db = new Client({
    host: config.dbHost,
    user: config.dbUser,
    password: config.dbPass,
    database: config.dbDatabase,
    port: config.dbPort,
});

db.connect();

logger.info("setting up the express server");
const app = express();

app.use(errorLogger);
app.use(requestLogger);

/**
 * Health check handler. Should be used by liveness and readiness checks.
 */
app.get("/healthz", async (req, res) => {
    const { rows } = await db
        .query("select version from version order by applied desc limit 1")
        .catch((e) => logger.error(e));

    if (!rows) {
        res.status(500).send("no rows returned from database");
        return;
    }

    res.status(200).send(`version ${rows[0].version}`);
});

app.get("/users/:username/apps/recently-added", recentlyAddedHandler(db));
app.get("/users/:username/analyses/recent", recentAnalysesHandler(db));
app.get("/users/:username/analyses/running", runningAnalysesHandler(db));

app.get("/apps/public", publicAppsHandler(db));

/**
 * Start up the server on the configured port.
 */
app.listen(config.listenPort, (err) => {
    if (err) throw err;
    console.log(`> Ready on http://localhost:${config.listenPort}`);
});
