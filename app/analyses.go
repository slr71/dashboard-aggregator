package app

import (
	"net/http"

	"github.com/cyverse-de/dashboard-aggregator/apis"
	"github.com/labstack/echo/v4"
)

func (a *App) RecentAnalysesForUser(c echo.Context) error {
	log := log.WithField("context", "recent analyses for user")

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

	analysisAPI := apis.NewAnalysisAPI(a.appsURL)

	recentAnalyses, err := analysisAPI.RecentAnalyses(username, int(limit))
	if err != nil {
		log.Error(err)
		return err
	}

	if err = c.JSON(http.StatusOK, recentAnalyses); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (a *App) RunningAnalysesForUser(c echo.Context) error {
	log := log.WithField("context", "recent analyses for user")

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

	analysisAPI := apis.NewAnalysisAPI(a.appsURL)

	runningAnalyses, err := analysisAPI.RunningAnalyses(username, int(limit))
	if err != nil {
		log.Error(err)
		return err
	}

	if err = c.JSON(http.StatusOK, runningAnalyses); err != nil {
		log.Error(err)
		return err
	}

	return err
}
