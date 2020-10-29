/**
 * Utility functions
 *
 * @module util
 */

/**
 * Validates an interval to be used in a database query and return the interval. A falsey value for the interval is
 * interpreted to mean that no interval was specified.
 *
 * @param {*} db the database connection
 * @param {string} interval the interval to validate
 */
export const validateInterval = async (db, interval) => {
    if (!interval) {
        return null;
    }

    // Use the DBMS to validate the interval.
    try {
        await db.query("select CAST($1 AS interval)", [interval]);
    } catch (e) {
        throw new Error(`invalid interval: ${interval}`);
    }

    return interval;
};

/**
 * Validates a row limit to be used in a database query. The limit must be a non-negative integer.
 *
 * @param {string} limitStr the string representation of the limit
 */
export const validateLimit = (limitStr) => {
    if (!limitStr) {
        return null;
    }

    let limit = parseInt(limitStr, 10);
    if (isNaN(limit)) {
        throw new Error(`invalid row limit: ${limitStr}`);
    }

    if (limit < 0) {
        throw new Error("the row limit may not be negative.");
    }

    return limit;
};
