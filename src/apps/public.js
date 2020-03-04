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

const getQuery = (appIDs) => `
 SELECT a.id,
        a.name,
        a.description,
        a.wiki_url,
        a.integration_date,
        a.edited_date
   FROM apps a
  WHERE a.id in ( ${appIDs.map((_, index) => `$${index + 1}`).join(",")} )
    AND a.deleted = false
    AND a.disabled = false
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
        .then((data) => data.permissions.map((p) => p.resource.name));
};

const getHandler = (db) => async (req, res) => {
    try {
        const appIDs = await getPublicAppIDs().catch((e) => {
            throw e;
        });

        const q = getQuery(appIDs);

        const { rows } = await db.query(q, appIDs).catch((e) => {
            throw e;
        });

        if (!rows) {
            throw new Error("no rows returned");
        }

        res.status(200).json({ apps: rows });
    } catch (e) {
        logger.error(e.message);
        res.status(500).send(`error running query: ${e.message}`);
    }
};

export default getHandler;
