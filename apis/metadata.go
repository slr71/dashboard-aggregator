package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type MetadataAPI struct {
	metadataURL *url.URL
}

func NewMetadataAPI(metadataURL *url.URL) *MetadataAPI {
	return &MetadataAPI{
		metadataURL: metadataURL,
	}
}

type TargetIDs struct {
	TargetIDs []string `json:"target-ids"`
}

func (m *MetadataAPI) GetFilteredTargetIDs(username string, targetTypes []string, avus []map[string]string, targetIDs []string) ([]string, error) {
	u := fixUsername(username)

	fullURL := *m.metadataURL.JoinPath("avus", "filter-targets")
	q := fullURL.Query()
	q.Set("user", u)
	fullURL.RawQuery = q.Encode()

	body := map[string]interface{}{}
	body["target-types"] = targetTypes
	body["target-ids"] = targetIDs
	body["avus"] = avus

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fullURL.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := string(rb)
		return nil, fmt.Errorf("url %s; status code %d; msg %s", fullURL.String(), resp.StatusCode, msg)
	}

	var data TargetIDs

	if err = json.Unmarshal(rb, &data); err != nil {
		return nil, err
	}

	retval := data.TargetIDs
	return retval, nil
}
