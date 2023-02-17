package app

import (
	"net/http"

	"github.com/cyverse-de/dashboard-aggregator/apis"
	"github.com/cyverse-de/dashboard-aggregator/db"
	"github.com/labstack/echo/v4"
)

func (a *App) UserDashboardHandler(c echo.Context) error {
	var err error

	log := log.WithField("context", "user dashboard")

	ctx := c.Request().Context()

	username, err := normalizeUsername(c)
	if err != nil {
		log.Error(err)
		return err
	}

	startDateInterval := normalizeStartDateInterval(c)

	limit, err := normalizeLimit(c)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debug("getting instant launch items")
	ilAPI, err := apis.NewInstantLaunchesAPI(a.config)
	if err != nil {
		log.Error(err)
		return err
	}
	ilItems, err := ilAPI.PullItems(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done getting instant launch items")

	analysisAPI := apis.NewAnalysisAPI(a.appsURL)

	log.Debug("getting recent analyses")
	recentAnalyses, err := analysisAPI.RecentAnalyses(ctx, username, int(limit))
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done getting recent analyses")

	log.Debug("getting running analyses")
	runningAnalyses, err := analysisAPI.RunningAnalyses(ctx, username, int(limit))
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("done getting running analyses")

	publicAppIDs, err := a.publicAppIDs(ctx)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
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
		log.Error(err)
		return err
	}
	log.Debug("done getting recently used apps")

	featuredAppIDs, err := a.featuredAppIDs(ctx, username, publicAppIDs)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
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
		log.Error(err)
		return err
	}

	return nil
}

func (a *App) PublicAppsForUserHandler(c echo.Context) error {
	log := log.WithField("context", "public apps for user")

	ctx := c.Request().Context()

	username, err := normalizeUsername(c)
	if err != nil {
		log.Error(err)
		return err
	}

	log = log.WithField("user", username)

	limit, err := normalizeLimit(c)
	if err != nil {
		log.Error(err)
		return err
	}

	publicAppIDs, err := a.publicAppIDs(ctx)
	if err != nil {
		log.Error(err)
		return err
	}

	publicApps, err := a.db.PublicAppsQuery(
		ctx,
		username,
		a.config.Apps.FavoritesGroupIndex,
		publicAppIDs,
		db.WithQueryLimit(uint(limit)),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	if err = c.JSON(http.StatusOK, map[string][]db.App{
		"apps": publicApps,
	}); err != nil {
		log.Error(err)
		return err
	}

	return err
}

func (a *App) RecentAddedAppsForUserHandler(c echo.Context) error {
	log := log.WithField("context", "recently added apps for user")

	ctx := c.Request().Context()

	username, err := normalizeUsername(c)
	if err != nil {
		log.Error(err)
		return err
	}

	log = log.WithField("user", username)

	limit, err := normalizeLimit(c)
	if err != nil {
		log.Error(err)
		return err
	}

	publicAppIDs, err := a.publicAppIDs(ctx)
	if err != nil {
		log.Error(err)
		return err
	}

	recentlyAddedApps, err := a.db.RecentlyAddedApps(
		ctx,
		username,
		a.config.Apps.FavoritesGroupIndex,
		publicAppIDs,
		db.WithQueryLimit(uint(limit)),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	if err = c.JSON(http.StatusOK, map[string][]db.App{
		"apps": recentlyAddedApps,
	}); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (a *App) PopularFeaturedAppsForUserHandler(c echo.Context) error {
	log := log.WithField("context", "popular featured apps for user")

	ctx := c.Request().Context()

	username, err := normalizeUsername(c)
	if err != nil {
		log.Error(err)
		return err
	}

	log = log.WithField("user", username)

	limit, err := normalizeLimit(c)
	if err != nil {
		log.Error(err)
		return err
	}

	startDateInterval := normalizeStartDateInterval(c)

	publicAppIDs, err := a.publicAppIDs(ctx)
	if err != nil {
		log.Error(err)
		return err
	}

	featuredAppIDs, err := a.featuredAppIDs(ctx, username, publicAppIDs)
	if err != nil {
		log.Error(err)
		return err
	}

	featuredApps, err := a.db.PopularFeaturedApps(
		ctx,
		&db.AppsQueryConfig{
			Username:          username,
			GroupsIndex:       a.config.Apps.FavoritesGroupIndex,
			AppIDs:            featuredAppIDs,
			StartDateInterval: startDateInterval,
		},
		db.WithQueryLimit(uint(limit)),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	if err = c.JSON(http.StatusOK, map[string][]db.App{
		"apps": featuredApps,
	}); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (a *App) RecentlyUsedAppsForUser(c echo.Context) error {
	log := log.WithField("context", "recently used apps for user")

	ctx := c.Request().Context()

	username, err := normalizeUsername(c)
	if err != nil {
		log.Error(err)
		return err
	}

	log = log.WithField("user", username)

	limit, err := normalizeLimit(c)
	if err != nil {
		log.Error(err)
		return err
	}

	startDateInterval := normalizeStartDateInterval(c)

	publicAppIDs, err := a.publicAppIDs(ctx)
	if err != nil {
		log.Error(err)
		return err
	}

	recentlyUsedApps, err := a.db.RecentlyUsedApps(
		ctx,
		&db.AppsQueryConfig{
			Username:          username,
			GroupsIndex:       a.config.Apps.FavoritesGroupIndex,
			AppIDs:            publicAppIDs,
			StartDateInterval: startDateInterval,
		},
		db.WithQueryLimit(uint(limit)),
	)
	if err != nil {
		log.Error(err)
		return err
	}

	if err = c.JSON(http.StatusOK, map[string][]db.App{
		"apps": recentlyUsedApps,
	}); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
