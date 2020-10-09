/**
 * @author johnworth
 *
 * Gets the list of public apps.
 *
 * @module apps/public
 */

import * as config from "../configuration";
import logger from "../logging";
import fetch from "node-fetch";

// All apps returned by this query are DE apps, so the system ID can be constant.
const getQuery = (appIDs) => `
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
            WHERE authenticated_user.username = $2
            AND acg.child_index = $3
            AND aca.app_id = a.id
         ) AS is_favorite
   FROM apps a
   JOIN integration_data d on a.integration_data_id = d.id
   JOIN users u on d.user_id = u.id
  WHERE a.id in ( ${appIDs.map((_, index) => `$${index + 4}`).join(",")} )
    AND a.deleted = false
    AND a.disabled = false
    AND a.integration_date IS NOT NULL
ORDER BY a.integration_date DESC
 LIMIT $1
`;

const getPublicAppIDs = () => {
    const reqURL = new URL(config.permissionsURL);
    reqURL.pathname = `/permissions/subjects/group/${config.publicGroup}/app`;
    return fetch(reqURL)
        .then(async (resp) => {
            if (!resp.ok) {
                const msg = await resp.text();
                throw new Error(msg);
            }
            return resp;
        })
        .then((resp) => resp.json())
        .then((data) => data.permissions.map((p) => p.resource.name))
        .catch((e) => {
            throw e;
        });
};

export const getData = async (db, username, limit) => {
    const appIDs = await getPublicAppIDs();

    const q = getQuery(appIDs);

    const { rows } = await db
        .query(q, [limit, username, config.favoritesGroupIndex, ...appIDs])
        .catch((e) => {
            throw e;
        });

    if (!rows) {
        throw new Error("no rows returned");
    }

    return rows;
};

const getHandler = (db) => async (req, res) => {
    try {
        // the parseInt isn't necessary, but it'll throw an error if the value
        // isn't a number.
        const username = req.params.username;
        const limit = parseInt(req?.query?.limit ?? "10", 10);
        const rows = await getData(db, username, limit);
        res.status(200).json({ apps: rows });
    } catch (e) {
        logger.error(e.message);
        res.status(500).send(`error running query: ${e.message}`);
    }
};

export default getHandler;
