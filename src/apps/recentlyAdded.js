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
         ) AS is_favorite
    FROM apps a
    JOIN integration_data i ON a.integration_data_id = i.id
    JOIN users u ON i.user_id = u.id
   WHERE a.deleted = false
     AND a.disabled = false
     AND u.username = $1
ORDER BY a.integration_date DESC
   LIMIT $2
 `;

export const getData = async (db, username, limit) => {
    const { rows } = await db
        .query(appsQuery, [username, limit, config.favoritesGroupIndex])
        .catch((e) => {
            throw e;
        });

    if (!rows) {
        throw new Error("no rows returned");
    }

    // Add the is_public flag to each app before returning the listing.
    const publicAppIds = new Set(await getPublicAppIDs());
    return rows.map((app) => ({ ...app, is_public: publicAppIds.has(app.id) }));
};

const getHandler = (db) => {
    return async (req, res) => {
        try {
            const username = req.params.username;
            const limit = validateLimit(req?.query?.limit) ?? 10;
            const rows = await getData(db, username, limit);
            res.status(200).json({ apps: rows });
        } catch (e) {
            logger.error(e.message);
            res.status(500).send(`error running query: ${e.message}`);
        }
    };
};

export default getHandler;
