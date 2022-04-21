/**
 * @author johnworth
 *
 * Gets the list of public apps.
 *
 * @module apps/public
 */

import { getPublicAppIDs } from "../clients/permissions";
import { validateLimit } from "../util";
import * as config from "../configuration";
import logger from "../logging";

import opentelemetry from "@opentelemetry/api";

function tracer() {
    return opentelemetry.trace.getTracer("dashboard-aggregator");
}

// All apps returned by this query are DE apps, so the system ID can be constant.
const publicAppsQuery = `
 SELECT a.id,
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
         true AS is_public
   FROM apps a
   JOIN integration_data d on a.integration_data_id = d.id
   JOIN users u on d.user_id = u.id
  WHERE a.id = ANY ($3)
    AND a.deleted = false
    AND a.disabled = false
    AND a.integration_date IS NOT NULL
ORDER BY a.integration_date DESC
 LIMIT $4
`;

export const getData = async (db, username, limit, publicAppIDs) => {
    const span = tracer().startSpan("apps/public getData");
    try {
        const { rows } = await db
            .query(publicAppsQuery, [
                username,
                config.favoritesGroupIndex,
                publicAppIDs,
                limit,
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

const getHandler = (db) => async (req, res) => {
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

export default getHandler;
