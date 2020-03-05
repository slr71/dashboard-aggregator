/**
 * @author johnworth
 *
 * Gathers information about apps recently added by the user.
 *
 * @module apps/recentlyAdded
 */

import logger from "../logging";

const appsQuery = `
  SELECT a.id,
         a.name,
         a.description,
         a.wiki_url,
         a.integration_date,
         a.edited_date
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
    const { rows } = await db.query(appsQuery, [username, limit]).catch((e) => {
        throw e;
    });

    if (!rows) {
        throw new Error("no rows returned");
    }

    return rows;
};

const getHandler = (db) => {
    return async (req, res) => {
        try {
            const username = req.params.username;
            const limit = parseInt(req?.query?.limit ?? "10", 10);
            const rows = await getData(db, username, limit);
            res.status(200).json({ apps: rows });
        } catch (e) {
            logger.error(e.message);
            res.status(500).send(`error running query: ${e.message}`);
        }
    };
};

export default getHandler;
