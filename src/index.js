import express from "express";
import { Client } from "node-postgres";

import * as config from "./configuration";
import logger, { errorLogger, requestLogger } from "./logging";

logger.info("creating database client");

logger.debug(config.dbURI);
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

app.listen(config.listenPort, (err) => {
    if (err) throw err;
    console.log(`> Ready on http://localhost:${config.listenPort}`);
});
