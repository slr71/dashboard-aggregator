/**
 *
 */

import * as config from "../configuration";
import fetch from "node-fetch";

export const getPublicAppIDs = async () => {
    const reqURL = new URL(config.permissionsURL);
    reqURL.pathname = `/permissions/subjects/group/${config.publicGroup}/app`;
    const resp = await fetch(reqURL);
    if (!resp.ok) {
        const msg = await resp.text();
        throw new Error(msg);
    }
    const data = await resp.json();
    return data.permissions.map((p) => p.resource.name);
};
