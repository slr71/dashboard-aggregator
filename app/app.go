package app

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/cyverse-de/dashboard-aggregator/apis"
	"github.com/cyverse-de/dashboard-aggregator/config"
	"github.com/cyverse-de/dashboard-aggregator/db"
	"github.com/cyverse-de/dashboard-aggregator/feeds"
	"github.com/cyverse-de/go-mod/httperror"
	"github.com/cyverse-de/go-mod/logging"
	"github.com/labstack/echo/v4"
)

var log = logging.Log.WithField("package", "app")

func normalizeLimit(c echo.Context) (int, error) {
	var (
		limit int64
		err   error
	)
	limit = DefaultLimit
	limitStr := c.QueryParam("limit")
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			return -1, echo.NewHTTPError(http.StatusBadRequest, "could not parse limit as an integer")
		}
	}
	return int(limit), nil
}

func normalizeStartDateInterval(c echo.Context) string {
	startDateInterval := c.QueryParam("start-date-interval")
	if startDateInterval == "" {
		startDateInterval = DefaultStartDateInterval
	}
	return startDateInterval
}

func normalizeUsername(c echo.Context) (string, error) {
	username := c.Param("username")
	if username == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, "username must be in the requested path")
	}
	return username, nil
}

type App struct {
	db             *db.Database
	ec             *echo.Echo
	pf             *feeds.PublicFeeds
	ilFeedURL      *url.URL
	appsURL        *url.URL
	metadataURL    *url.URL
	permissionsURL *url.URL
	config         *config.ServiceConfiguration
}

func New(db *db.Database, pf *feeds.PublicFeeds, cfg *config.ServiceConfiguration) (*App, error) {
	ilURL, err := url.Parse(cfg.AppExposer.URL)
	if err != nil {
		return nil, err
	}
	appsURL, err := url.Parse(cfg.Apps.URL)
	if err != nil {
		return nil, err
	}
	metadataURL, err := url.Parse(cfg.Metadata.URL)
	if err != nil {
		return nil, err
	}
	permissionsURL, err := url.Parse(cfg.Permissions.URL)
	if err != nil {
		return nil, err
	}
	return &App{
		db:             db,
		ec:             echo.New(),
		pf:             pf,
		ilFeedURL:      ilURL,
		appsURL:        appsURL,
		metadataURL:    metadataURL,
		permissionsURL: permissionsURL,
		config:         cfg,
	}, nil
}

func (a *App) Echo() *echo.Echo {
	a.ec.HTTPErrorHandler = httperror.HTTPErrorHandler

	a.ec.GET("/", a.HelloHandler)
	a.ec.GET("/feeds", a.PublicFeedsHandler)

	users := a.ec.Group("/users")
	users.GET("/:username", a.UserDashboardHandler)
	return a.ec
}

func (a *App) HelloHandler(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Hello from dashboard-aggregator")
}

func (a *App) PublicFeedsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	result := a.pf.Marshallable(ctx)
	return c.JSON(http.StatusOK, &result)
}

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

	permissionsAPI := apis.NewPermissionsAPI(a.permissionsURL)

	log.Debug("getting public app ids")
	publicAppIDs, err := permissionsAPI.GetPublicIDS(a.config.Permissions.PublicGroup)
	if err != nil {
		return err
	}
	log.Debug("done getting public app ids")

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

	metadataAPI := apis.NewMetadataAPI(a.metadataURL)

	featuredAppsAVUs := []map[string]string{
		{
			"attr":  a.config.Metadata.FeaturedAppsAttribute,
			"value": a.config.Metadata.FeaturedAppsValue,
		},
	}

	log.Debug("getting featured app ids")
	featuredAppIDs, err := metadataAPI.GetFilteredTargetIDs(username, []string{"app"}, featuredAppsAVUs, publicAppIDs)
	if err != nil {
		return err
	}
	log.Debug("done getting featured app ids")

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
