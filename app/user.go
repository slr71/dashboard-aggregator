package app

import (
	"net/http"

	"github.com/cyverse-de/dashboard-aggregator/apis"
	"github.com/cyverse-de/dashboard-aggregator/db"
	"github.com/labstack/echo/v4"
)

func (a *App) UserDashboardHandler(c echo.Context) error {
	var err error

	ctx := c.Request().Context()

	username, err := normalizeUsername(c)
	if err != nil {
		return err
	}

	startDateInterval := normalizeStartDateInterval(c)

	limit, err := normalizeLimit(c)
	if err != nil {
		return err
	}

	log.Debug("getting instant launch items")
	ilAPI, err := apis.NewInstantLaunchesAPI(a.config)
	if err != nil {
		return err
	}
	ilItems, err := ilAPI.PullItems(ctx)
	if err != nil {
		return err
	}
	log.Debug("done getting instant launch items")

	analysisAPI := apis.NewAnalysisAPI(a.appsURL)

	log.Debug("getting recent analyses")
	recentAnalyses, err := analysisAPI.RecentAnalyses(username, int(limit))
	if err != nil {
		return err
	}
	log.Debug("done getting recent analyses")

	log.Debug("getting running analyses")
	runningAnalyses, err := analysisAPI.RunningAnalyses(username, int(limit))
	if err != nil {
		return err
	}
	log.Debug("done getting running analyses")

	publicAppIDs, err := a.publicAppIDs()
	if err != nil {
		return err
	}

	log.Debug("getting recently added apps")
	recentlyAddedApps, err := a.db.RecentlyAddedApps(ctx, username, a.config.Apps.FavoritesGroupIndex, publicAppIDs, db.WithQueryLimit(uint(limit)))
	if err != nil {
		return err
	}
	log.Debug("done getting recently added apps")

	log.Debug("getting public apps")
	publicApps, err := a.db.PublicAppsQuery(ctx, username, a.config.Apps.FavoritesGroupIndex, publicAppIDs, db.WithQueryLimit(uint(limit)))
	if err != nil {
		return err
	}
	log.Debug("done getting public apps")

	log.Debug("getting recently used apps")
	recentlyUsed, err := a.db.RecentlyUsedApps(ctx, &db.AppsQueryConfig{
		Username:          username,
		GroupsIndex:       a.config.Apps.FavoritesGroupIndex,
		AppIDs:            publicAppIDs,
		StartDateInterval: startDateInterval,
	}, db.WithQueryLimit(uint(limit)))
	if err != nil {
		return err
	}
	log.Debug("done getting recently used apps")

	featuredAppIDs, err := a.featuredAppIDs(username, publicAppIDs)
	if err != nil {
		return err
	}

	log.Debug("getting featured apps")
	featuredApps, err := a.db.PopularFeaturedApps(ctx, &db.AppsQueryConfig{
		Username:          username,
		GroupsIndex:       a.config.Apps.FavoritesGroupIndex,
		AppIDs:            featuredAppIDs,
		StartDateInterval: startDateInterval,
	}, db.WithQueryLimit(uint(limit)))
	if err != nil {
		return err
	}
	log.Debug("done getting featured apps")

	publicFeeds := a.pf

	retval := map[string]interface{}{
		"analyses": map[string]interface{}{
			"recent":  recentAnalyses.Analyses,
			"running": runningAnalyses.Analyses,
		},
		"apps": map[string]interface{}{
			"recentlyAdded":   recentlyAddedApps,
			"public":          publicApps,
			"recentlyUsed":    recentlyUsed,
			"popularFeatured": featuredApps,
		},
		"instantLaunches": ilItems,
		"feeds":           publicFeeds.Marshallable(ctx),
	}

	if err = c.JSON(http.StatusOK, retval); err != nil {
		return err
	}

	return nil
}
