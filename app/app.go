package app

import (
	"net/http"

	"github.com/cyverse-de/dashboard-aggregator/db"
	"github.com/cyverse-de/dashboard-aggregator/feeds"
	"github.com/cyverse-de/go-mod/httperror"
	"github.com/cyverse-de/go-mod/logging"
	"github.com/labstack/echo/v4"
)

var log = logging.Log.WithField("package", "app")

type App struct {
	db *db.Database
	ec *echo.Echo
	pf *feeds.PublicFeeds
}

func New(db *db.Database, pf *feeds.PublicFeeds) *App {
	return &App{
		db: db,
		ec: echo.New(),
		pf: pf,
	}
}

func (a *App) Echo() *echo.Echo {
	a.ec.HTTPErrorHandler = httperror.HTTPErrorHandler

	a.ec.GET("/", a.HelloHandler)
	a.ec.GET("/feeds", a.PublicFeedsHandler)

	// users := a.ec.Group("/users")
	// users.GET("/:username", a.UserDashboardHandler)
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
