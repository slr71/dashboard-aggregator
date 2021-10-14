/**
 * Takes a list of ids and sends a request to the metadata
 * service to filter those IDs to just those with the corresponding AVUs and
 * target types
 */

import * as config from "../configuration";
import fetch from "node-fetch";

/**
 * @param username - The username of the authenticated user
 * @param targetTypes - Array of the types of resources to include
 * @param avus - Array of the metadata avus to use for filtering
 * @param targetIds - Array of the resource IDs to include and filter
 * @returns {Promise<*>}
 */
export const getFilteredTargetIds = async ({
    username,
    targetTypes,
    avus,
    targetIds,
}) => {
    const reqURL = new URL(config.metadataURL);
    reqURL.pathname = "/avus/filter-targets";
    reqURL.search = new URLSearchParams({ user: username }).toString();
    const body = {
        "target-types": targetTypes,
        "target-ids": targetIds,
        avus,
    };

    const resp = await fetch(reqURL, {
        method: "post",
        body: JSON.stringify(body),
        headers: { "Content-Type": "application/json" },
    });

    if (!resp.ok) {
        const msg = await resp.text();
        throw new Error(msg);
    }

    const data = await resp.json();
    return data["target-ids"];
};
