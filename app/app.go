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

const DefaultStartDateInterval = "1 year"
const DefaultLimit = int64(10)

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

	a.ec.GET("/", a.LoggedOutHandler)
	a.ec.GET("/hello", a.HelloHandler)
	a.ec.GET("/feeds", a.PublicFeedsHandler)

	users := a.ec.Group("/users")
	users.GET("/:username", a.UserDashboardHandler)

	apps := a.ec.Group("/apps")
	apps.GET("/public", a.PublicAppsHandler)
	apps.GET("/recently-ran", a.RecentlyRunAppsHandler)

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

func (a *App) featuredAppIDs(username string, publicAppIDs []string) ([]string, error) {
	log := log.WithField("context", "featured app ids lookup")

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
		return nil, err
	}
	log.Debug("done getting featured app ids")

	return featuredAppIDs, nil
}

func (a *App) publicAppIDs() ([]string, error) {
	log := log.WithField("context", "public app ids lookup")

	permissionsAPI := apis.NewPermissionsAPI(a.permissionsURL)

	log.Debug("getting public app ids")
	publicAppIDs, err := permissionsAPI.GetPublicIDS(a.config.Permissions.PublicGroup)
	if err != nil {
		return nil, err
	}
	log.Debug("done getting public app ids")

	return publicAppIDs, nil
}
