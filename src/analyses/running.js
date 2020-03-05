/**
 * @author johnworth
 *
 * Returns a list of recently used apps.
 *
 * @module apps/running
 */

import logger from "../logging";

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
         j.planned_end_date,
         j.status,
         j.subdomain,
         j.parent_id
    FROM jobs j
    JOIN users ON j.user_id = users.id
   WHERE j.deleted = false
     AND j.status = 'Running'
     AND users.username = $1
ORDER BY start_date DESC
   LIMIT $2
`;

const getHandler = (db) => async (req, res) => {
    try {
        // The parseInt should raise an error if it fails.
        const limit = parseInt(req?.query?.limit ?? "10", 10);
        const username = req?.params?.username;

        const { rows } = await db
            .query(recentAnalyses, [username, limit])
            .catch((e) => {
                throw e;
            });

        if (!rows) {
            throw new Error("no rows returned");
        }

        res.status(200).json({ analyses: rows });
    } catch (e) {
        logger.error(e);
        res.status(500).send(e.message);
    }
};

export default getHandler;
