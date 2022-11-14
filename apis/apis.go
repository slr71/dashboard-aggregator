package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cyverse-de/go-mod/logging"
)

var log = logging.Log.WithField("package", "apis")

type Rows map[string]interface{}

type AnalysisAPI struct {
	appsURL *url.URL
}

func fixUsername(username string) string {
	parts := strings.Split(username, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return username
}

func (a *AnalysisAPI) RunningAnalyses(username string, limit int) (Rows, error) {
	log := log.WithField("context", "running analyses")

	u := fixUsername(username)
	log = log.WithField("user", u)

	fullURL := *a.appsURL

	filter := []map[string]string{
		{
			"field": "status",
			"value": "Running",
		},
	}

	filterStr, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	q := fullURL.Query()
	q.Set("limit", strconv.FormatInt(int64(limit), 10))
	q.Set("user", u)
	q.Set("filter", string(filterStr))
	fullURL.RawQuery = q.Encode()

	log.Infof("getting running analyses", u)

	resp, err := http.Get(fullURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Infof("done getting running analyses", u)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code from %s was %d", fullURL.String(), resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data Rows
	if err = json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (a *AnalysisAPI) RecentAnalyses(username string, limit int) (Rows, error) {
	log := log.WithField("context", "recent analyses")

	u := fixUsername(username)
	log = log.WithField("user", u)

	fullURL := a.appsURL

	q := fullURL.Query()
	q.Set("limit", strconv.FormatInt(int64(limit), 10))
	q.Set("user", u)
	q.Set("sort-field", "startdate")
	q.Set("sort-dir", "DESC")
	fullURL.RawQuery = q.Encode()

	log.Infof("getting recent analyses", u)

	resp, err := http.Get(fullURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Infof("done getting recent analyses", u)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code from %s was %d", fullURL.String(), resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data Rows
	if err = json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return data, nil

}
