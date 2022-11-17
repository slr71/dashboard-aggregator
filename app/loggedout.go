package app

import (
	"net/http"

	"github.com/cyverse-de/dashboard-aggregator/db"
	"github.com/labstack/echo/v4"
)

func (a *App) LoggedOutHandler(c echo.Context) error {
	ctx := c.Request().Context()

	limit, err := normalizeLimit(c)
	if err != nil {
		return err
	}

	username := "anonymous"

	startDateInterval := normalizeStartDateInterval(c)

	feeds := a.pf.Marshallable(ctx)

	publicAppIDs, err := a.publicAppIDs()
	if err != nil {
		return err
	}

	featuredAppIDs, err := a.featuredAppIDs(username, publicAppIDs)
	if err != nil {
		return err
	}

	popularFeaturedApps, err := a.db.PopularFeaturedApps(ctx, &db.AppsQueryConfig{
		Username:          username,
		GroupsIndex:       a.config.Apps.FavoritesGroupIndex,
		AppIDs:            featuredAppIDs,
		StartDateInterval: startDateInterval,
	}, db.WithQueryLimit(uint(limit)))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"apps": map[string][]db.App{
			"popularFeatured": popularFeaturedApps,
		},
		"feeds": feeds,
	})
}
