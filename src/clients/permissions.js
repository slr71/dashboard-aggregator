/**
 *
 */

import * as config from "../configuration";
import fetch from "node-fetch";

import opentelemetry from "@opentelemetry/api";

function tracer() {
    return opentelemetry.trace.getTracer("dashboard-aggregator");
}

export const getPublicAppIDs = async () => {
    return tracer().startActiveSpan("getPublicAppIDs", async (span) => {
        try {
            const reqURL = new URL(config.permissionsURL);
            reqURL.pathname = `/permissions/abbreviated/subjects/group/${config.publicGroup}/app`;
            const resp = await fetch(reqURL);
            if (!resp.ok) {
                const msg = await resp.text();
                throw new Error(msg);
            }
            const data = await resp.json();
            return data.permissions.map((p) => p.resource_name);
        } finally {
            span.end();
        }
    });
};
