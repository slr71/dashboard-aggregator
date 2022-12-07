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

type AnalysisListing struct {
	Analyses []interface{} `json:"analyses"`
}

type AnalysisAPI struct {
	appsURL *url.URL
}

func NewAnalysisAPI(appsURL *url.URL) *AnalysisAPI {
	return &AnalysisAPI{
		appsURL: appsURL,
	}
}

func fixUsername(username string) string {
	parts := strings.Split(username, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return username
}

func (a *AnalysisAPI) RunningAnalyses(username string, limit int) (*AnalysisListing, error) {
	log := log.WithField("context", "running analyses")

	u := fixUsername(username)
	log = log.WithField("user", u)

	fullURL := *a.appsURL.JoinPath("analyses")

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

	fullURL.RawQuery = fmt.Sprintf("%s&filter=%s", q.Encode(), string(filterStr))

	log.Debugf("getting running analyses from %s", fullURL.String())

	resp, err := http.Get(fullURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debugf("done getting running analyses from %s", fullURL.String())

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code from %s was %d", fullURL.String(), resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data AnalysisListing
	if err = json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (a *AnalysisAPI) RecentAnalyses(username string, limit int) (*AnalysisListing, error) {
	log := log.WithField("context", "recent analyses")

	u := fixUsername(username)
	log = log.WithField("user", u)

	fullURL := a.appsURL.JoinPath("analyses")

	q := fullURL.Query()
	q.Set("limit", strconv.FormatInt(int64(limit), 10))
	q.Set("user", u)
	q.Set("sort-field", "startdate")
	q.Set("sort-dir", "DESC")
	fullURL.RawQuery = q.Encode()

	log.Debugf("getting recent analyses from %s", fullURL.String())

	resp, err := http.Get(fullURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debugf("done getting recent analyses from %s", fullURL.String())

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := string(b)
		return nil, fmt.Errorf("url %s; status code %d; msg: %s", fullURL.String(), resp.StatusCode, msg)
	}

	var data AnalysisListing
	if err = json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return &data, nil

}
