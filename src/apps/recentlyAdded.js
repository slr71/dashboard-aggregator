/**
 * @author johnworth
 *
 * Gathers information about apps recently added by the user.
 *
 * @module apps/recentlyAdded
 */

import { getPublicAppIDs } from "../clients/permissions";
import * as config from "../configuration";
import logger from "../logging";
import { validateLimit } from "../util";

import opentelemetry from "@opentelemetry/api";

function tracer() {
    return opentelemetry.trace.getTracer("dashboard-aggregator");
}

// All apps returned by this query are DE apps, so the system ID can be constant.
const appsQuery = `
  SELECT a.id,
         'de' AS system_id,
         a.name,
         a.description,
         a.wiki_url,
         a.integration_date,
         a.edited_date,
         u.username,
         EXISTS (
            SELECT * FROM workspace w
            JOIN app_category_group acg ON w.root_category_id = acg.parent_category_id
            JOIN app_category_app aca ON acg.child_category_id = aca.app_category_id
            WHERE aca.app_id = a.id
            AND w.user_id = u.id
            AND acg.child_index = $3
         ) AS is_favorite,
         (a.id = ANY ($4)) AS is_public
    FROM apps a
    JOIN integration_data i ON a.integration_data_id = i.id
    JOIN users u ON i.user_id = u.id
   WHERE a.deleted = false
     AND a.disabled = false
     AND u.username = $1
ORDER BY a.integration_date DESC
   LIMIT $2
 `;

export const getData = async (db, username, limit, publicAppIDs) => {
    const span = tracer().startSpan("apps/recentlyAdded getData");
    try {
        const { rows } = await db
            .query(appsQuery, [
                username,
                limit,
                config.favoritesGroupIndex,
                publicAppIDs,
            ])
            .catch((e) => {
                throw e;
            });

        if (!rows) {
            throw new Error("no rows returned");
        }

        return rows;
    } finally {
        span.end();
    }
};

const getHandler = (db) => {
    return async (req, res) => {
        try {
            const username = req.params.username;
            const limit = validateLimit(req?.query?.limit) ?? 10;
            const publicAppIDs = await getPublicAppIDs();
            const rows = await getData(db, username, limit, publicAppIDs);
            res.status(200).json({ apps: rows });
        } catch (e) {
            logger.error(e.message);
            res.status(500).send(`error running query: ${e.message}`);
        }
    };
};

export default getHandler;
