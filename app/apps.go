package app

import (
	"net/http"

	"github.com/cyverse-de/dashboard-aggregator/db"
	"github.com/labstack/echo/v4"
)

func (a *App) PublicAppsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	limit, err := normalizeLimit(c)
	if err != nil {
		log.Error(err)
		return err
	}

	publicAppIDs, err := a.publicAppIDs()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug("getting public apps")
	publicApps, err := a.db.PublicAppsQuery(ctx, "", a.config.Apps.FavoritesGroupIndex, publicAppIDs, db.WithQueryLimit(uint(limit)))
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done getting public apps")

	if err = c.JSON(http.StatusOK, map[string][]db.App{
		"apps": publicApps,
	}); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (a *App) RecentlyRunAppsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	limit, err := normalizeLimit(c)
	if err != nil {
		log.Error(err)
		return err
	}

	startDateInterval := normalizeStartDateInterval(c)

	publicAppIDs, err := a.publicAppIDs()
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug("getting recently used apps")
	recentlyUsedApps, err := a.db.RecentlyUsedApps(ctx, &db.AppsQueryConfig{
		Username:          "",
		GroupsIndex:       a.config.Apps.FavoritesGroupIndex,
		AppIDs:            publicAppIDs,
		StartDateInterval: startDateInterval,
	}, db.WithQueryLimit(uint(limit)))
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done getting recently used apps")

	if err = c.JSON(http.StatusOK, map[string][]db.App{
		"apps": recentlyUsedApps,
	}); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
