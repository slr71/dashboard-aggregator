/**
 * @author aramsey
 *
 * Returns the list of most popular Featured apps (aka certified apps or blessed
 * apps)
 */

import { getPublicAppIDs } from "../clients/permissions";
import { getFilteredTargetIds } from "../clients/metadata";

import * as config from "../configuration";
import logger from "../logging";
import { validateInterval, validateLimit } from "../util";
import constants from "../constants";

import opentelemetry from "@opentelemetry/api";

function tracer() {
    return opentelemetry.trace.getTracer("dashboard-aggregator");
}

const popularFeaturedAppsQuery = `
    SELECT a.id,
           'de'        AS system_id,
           a.name,
           a.description,
           a.wiki_url,
           a.integration_date,
           a.edited_date,
           a.integrator_username AS username,
           count(j.id) AS job_count,
           EXISTS(
                   SELECT *
                   FROM users authenticated_user
                            JOIN workspace w ON authenticated_user.id = w.user_id
                            JOIN app_category_group acg ON w.root_category_id = acg.parent_category_id
                            JOIN app_category_app aca ON acg.child_category_id = aca.app_category_id
                   WHERE authenticated_user.username = $1
                     AND acg.child_index = $3
                     AND aca.app_id = a.id
               )       AS is_favorite,
           true        AS is_public
    FROM app_listing a
             LEFT JOIN jobs j on j.app_id = CAST(a.id as TEXT)
    WHERE a.id = ANY ($4)
      AND a.deleted = false
      AND a.disabled = false
      AND a.integration_date IS NOT NULL
      AND (j.start_date >= (now() - CAST($5 AS interval)) OR j.start_date IS NULL)
    GROUP BY a.id, a.name, a.description, a.wiki_url, a.integration_date, a.edited_date, a.integrator_username
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
    return tracer().startActiveSpan(
        "apps/popularFeatured getData",
        async (span) => {
            try {
                const { rows } = await db
                    .query(popularFeaturedAppsQuery, [
                        username,
                        limit,
                        config.favoritesGroupIndex,
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
            } finally {
                span.end();
            }
        }
    );
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
        res.status(500).json({ reason: e.message });
    }
};

export default getHandler;
