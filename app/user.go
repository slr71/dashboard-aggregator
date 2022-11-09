package app

const DefaultStartDateInterval = "1 year"
const DefaultLimit = int64(10)

// func (a *App) UserDashboardHandler(c echo.Context) error {
// 	var err error

// 	ctx := c.Request().Context()

// 	username := c.Param("username")
// 	if username == "" {
// 		return echo.NewHTTPError(http.StatusBadRequest, "missing username from request")
// 	}

// 	startDateInterval := c.QueryParam("start-date-interval")
// 	if startDateInterval == "" {
// 		startDateInterval = DefaultStartDateInterval
// 	}
// 	if err = a.db.ValidateInterval(ctx, startDateInterval); err != nil {
// 		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 	}

// 	limit := DefaultLimit
// 	limitStr := c.QueryParam("limit")
// 	if limitStr != "" {
// 		limit, err = strconv.ParseInt(limitStr, 10, 32)
// 		if err != nil {
// 			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
// 		}
// 	}

// 	return nil
// }
