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
         a.integrator_username AS username,
         EXISTS (
            SELECT *
            FROM users authenticated_user
                     JOIN workspace w ON authenticated_user.id = w.user_id
                     JOIN app_category_group acg ON w.root_category_id = acg.parent_category_id
                     JOIN app_category_app aca ON acg.child_category_id = aca.app_category_id
            WHERE authenticated_user.username = $1
              AND acg.child_index = $3
              AND aca.app_id = a.id
         ) AS is_favorite,
         (a.id = ANY ($4)) AS is_public
    FROM app_listing a
   WHERE a.deleted = false
     AND a.disabled = false
     AND a.integrator_username = $1
ORDER BY a.integration_date DESC
   LIMIT $2
 `;

export const getData = async (db, username, limit, publicAppIDs) => {
    return tracer().startActiveSpan(
        "apps/recentlyAdded getData",
        async (span) => {
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
        }
    );
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
            res.status(500).json({
                reason: `error running query: ${e.message}`,
            });
        }
    };
};

export default getHandler;
