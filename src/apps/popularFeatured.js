/**
 * @author aramsey
 *
 * Returns the list of most popular Featured apps (aka certified apps or blessed
 * apps)
 */

import { getPublicAppIDs } from "../clients/permissions";
import { getFilteredTargetIds } from "../clients/metadata";
import logger from "../logging";
import { validateInterval, validateLimit } from "../util";
import constants from "../constants";

const popularFeaturedAppsQuery = `
    SELECT a.id,
           'de'        AS system_id,
           a.name,
           a.description,
           a.wiki_url,
           a.integration_date,
           a.edited_date,
           u.username,
           count(j.id) AS job_count,
           EXISTS(
                   SELECT *
                   FROM users authenticated_user
                            JOIN workspace w ON authenticated_user.id = w.user_id
                            JOIN app_category_group acg ON w.root_category_id = acg.parent_category_id
                            JOIN app_category_app aca ON acg.child_category_id = aca.app_category_id
                   WHERE authenticated_user.username = $1
                     AND acg.child_index = $2
                     AND aca.app_id = a.id
               )       AS is_favorite,
           true        AS is_public
    FROM apps a
             JOIN jobs j on j.app_id = CAST(a.id as TEXT)
             JOIN integration_data d on a.integration_data_id = d.id
             JOIN users u on d.user_id = u.id
    WHERE a.id = ANY ($3)
      AND a.deleted = false
      AND a.disabled = false
      AND a.integration_date IS NOT NULL
      AND j.start_date >= (now() - CAST($4 AS interval))
    GROUP BY a.id, u.id
    ORDER BY job_count DESC
        LIMIT $2
`;

export const getData = async (
    db,
    username,
    limit,
    featuredAppIds,
    startDateInterval
) => {
    const { rows } = await db
        .query(popularFeaturedAppsQuery, [
            username,
            limit,
            featuredAppIds,
            startDateInterval,
        ])
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
        const publicAppIDs = await getPublicAppIDs();
        const username = req.params.username;
        const limit = validateLimit(req?.query?.limit) ?? 10;
        const startDateInterval =
            (await validateInterval(db, req?.query["start-date-interval"])) ??
            constants.DEFAULT_START_DATE_INTERVAL;
        const featuredAppIds = await getFilteredTargetIds({
            targetTypes: ["app"],
            targetIds: publicAppIDs,
            avus: constants.FEATURED_APPS_AVUS,
            username,
        });
        const rows = await getData(
            db,
            username,
            limit,
            featuredAppIds,
            startDateInterval
        );
        res.status(200).json({ apps: rows });
    } catch (e) {
        logger.error(e);
        res.status(500).send(e.message);
    }
};

export default getHandler;
