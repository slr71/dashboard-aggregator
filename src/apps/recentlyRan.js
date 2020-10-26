/**
 * Gets the list of apps that have ran recently.
 *
 * @module apps/recentlyRan
 */

import { getPublicAppIDs } from "../clients/permissions";
import { validateInterval, validateLimit } from "../util";
import * as config from "../configuration";
import logger from "../logging";

// All apps returned by this query are DE apps, so the system ID can be constant.
const recentlyRanAppsQuery = `
    SELECT DISTINCT
        a.id,
        'de' AS system_id,
        a.name,
        a.description,
        a.wiki_url,
        a.integration_date,
        a.edited_date,
        u.username,
        EXISTS (
            SELECT * FROM users authenticated_user
            JOIN workspace w ON authenticated_user.id = w.user_id
            JOIN app_category_group acg ON w.root_category_id = acg.parent_category_id
            JOIN app_category_app aca ON acg.child_category_id = aca.app_category_id
            WHERE authenticated_user.username = $1
            AND acg.child_index = $2
            AND aca.app_id = a.id
        ) AS is_favorite,
        TRUE AS is_public,
        max(j.start_date) AS most_recent_start_date
    FROM jobs j
    JOIN apps a on CAST(a.id AS text) = j.app_id
    JOIN integration_data d on a.integration_data_id = d.id
    JOIN users u on d.user_id = u.id
    WHERE a.id = ANY ($3)
    AND NOT a.deleted
    AND NOT a.disabled
    AND j.start_date > now() - CAST($4 AS interval)
    GROUP BY a.id, a.name, a.description, a.wiki_url, a.integration_date, a.edited_date, u.username
    ORDER BY most_recent_start_date DESC
    LIMIT $5
`;

export const getRecentlyRanApps = async (
    db,
    username,
    limit,
    startDateInterval
) => {
    const appIDs = await getPublicAppIDs();

    const { rows } = await db
        .query(recentlyRanAppsQuery, [
            username,
            config.favoritesGroupIndex,
            appIDs,
            startDateInterval,
            limit,
        ])
        .catch((e) => {
            logger.error(`something bad happened: ${e}`);
            throw e;
        });

    if (!rows) {
        throw new Error("no rows returned");
    }

    // Remove unwanted columns from the result; doing this in SQL made the query a little clunky.
    for (const r of rows) {
        delete r["most_recent_start_date"];
    }

    return rows;
};

const getHandler = (db) => async (req, res) => {
    try {
        const username = req.params.username;
        const limit = validateLimit(req?.query?.limit) ?? 10;
        const startDateInterval =
            (await validateInterval(db, req?.query["start-date-interval"])) ??
            "1 week";

        // Query the database.
        const rows = await getRecentlyRanApps(
            db,
            username,
            limit,
            startDateInterval
        );

        res.status(200).json({ apps: rows });
    } catch (e) {
        logger.error(e.message);
        res.status(500).send(`error running query: ${e.message}`);
    }
};

export default getHandler;
