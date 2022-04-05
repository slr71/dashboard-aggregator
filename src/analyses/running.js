/**
 * @author johnworth, sriram
 *
 * Returns a list of running analyses.
 *
 * @module analyses/running
 */

import logger from "../logging";
import { validateLimit } from "../util";
import axios from "axios";
import * as config from "../configuration";

import opentelemetry from "@opentelemetry/api";

function tracer() {
    return opentelemetry.trace.getTracer("dashboard-aggregator");
}

export const getData = async (username, limit) => {
    return tracer().startActiveSpan(
        "analyses/running getData",
        async (span) => {
            try {
                const { data } = await axios.get(
                    `${config.appsURL}/analyses?limit=${limit}&user=${
                        username?.split("@")[0]
                    }&filter=[{"field":"status", "value":"Running"}]`
                );
                logger.info(
                    "Running analyses for user " +
                        username +
                        ": " +
                        JSON.stringify(data)
                );
                return data;
            } catch (e) {
                span.setStatus({
                    code: opentelemetry.SpanStatusCode.ERROR,
                    message: e,
                });
                throw new Error(e);
            } finally {
                span.end();
            }
        }
    );
};

const getHandler = () => async (req, res) => {
    try {
        const limit = validateLimit(req?.query?.limit) ?? 10;
        const username = req?.params?.username?.split("@")[0];
        const rows = await getData(username, limit);
        res.status(200).json(rows);
    } catch (e) {
        logger.error(e);
        res.status(500).json({ reason: e.message });
    }
};

export default getHandler;
