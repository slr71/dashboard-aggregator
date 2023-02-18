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

	// Fetch instant launches
	ilAPI, err := apis.NewInstantLaunchesAPI(a.config)
	if err != nil {
		log.Error(err)
		return err
	}

	ilChan := make(chan []map[string]interface{})
	ilErrChan := make(chan error)

	go ilAPI.PullItemsAsync(ctx, ilChan, ilErrChan)

	// Fetch recent & running analyses
	analysisAPI := apis.NewAnalysisAPI(a.appsURL)

	recentAnalysisChan := make(chan *apis.AnalysisListing)
	recentAnalysisErrChan := make(chan error)

	runningAnalysisChan := make(chan *apis.AnalysisListing)
	runningAnalysisErrChan := make(chan error)

	go analysisAPI.RecentAnalysesAsync(ctx, recentAnalysisChan, recentAnalysisErrChan, username, int(limit))
	go analysisAPI.RunningAnalysesAsync(ctx, runningAnalysisChan, runningAnalysisErrChan, username, int(limit))

	// Fetch public & featured app IDs
	publicAppIDsChan := make(chan []string)
	publicAppIDsErrChan := make(chan error)

	go a.publicAppIDsAsync(ctx, publicAppIDsChan, publicAppIDsErrChan)

	// We need public app IDs for the next few calls
	err = <-publicAppIDsErrChan
	if err != nil {
		log.Error(err)
		return err
	}
	publicAppIDs := <-publicAppIDsChan

	featuredAppIDsChan := make(chan []string)
	featuredAppIDsErrChan := make(chan error)

	go a.featuredAppIDsAsync(ctx, featuredAppIDsChan, featuredAppIDsErrChan, username, publicAppIDs)

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

	// We need featured app IDs for the next bit
	err = <-featuredAppIDsErrChan
	if err != nil {
		log.Error(err)
		return err
	}
	featuredAppIDs := <-featuredAppIDsChan

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

	// Now, check all the channels we still haven't
	err = <-ilErrChan
	if err != nil {
		log.Error(err)
		return err
	}
	ilItems := <-ilChan

	err = <-recentAnalysisErrChan
	if err != nil {
		log.Error(err)
		return err
	}
	recentAnalyses := <-recentAnalysisChan

	err = <-runningAnalysisErrChan
	if err != nil {
		log.Error(err)
		return err
	}
	runningAnalyses := <-runningAnalysisChan

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
