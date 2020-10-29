/**
 * @author johnworth
 *
 * Returns a list of recent analyses.
 *
 * @module analyses/recent
 */

import logger from "../logging";
import { validateLimit } from "../util";

const recentAnalyses = `
  SELECT j.id,
         j.job_name as name,
         j.job_description as description,
         j.app_id,
         j.app_name,
         j.app_description,
         j.result_folder_path,
         j.start_date,
         j.end_date,
         j.planned_end_date ,
         j.status,
         j.subdomain,
         j.parent_id,
         u.username
    FROM jobs j
    JOIN users u ON j.user_id = u.id
   WHERE j.deleted = false
     AND u.username = $1
ORDER BY start_date DESC
   LIMIT $2
`;

export const getData = async (db, username, limit) => {
    const { rows } = await db
        .query(recentAnalyses, [username, limit])
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
        const limit = validateLimit(req?.query?.limit) ?? 10;
        const username = req?.params?.username;
        const rows = await getData(db, username, limit);
        res.status(200).json({ analyses: rows });
    } catch (e) {
        logger.error(e);
        res.status(500).send(e.message);
    }
};

export default getHandler;
