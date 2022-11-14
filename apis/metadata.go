package apis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type MetadataAPI struct {
	metadataURL *url.URL
}

type AVUs map[string]string

func (m *MetadataAPI) GetFilteredTargetIDS(username string, targetTypes []string, avus AVUs, targetIDs []string) ([]string, error) {
	u := fixUsername(username)

	fullURL := *m.metadataURL
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code from %s was %d", fullURL.String(), resp.StatusCode)
	}

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}

	if err = json.Unmarshal(rb, &data); err != nil {
		return nil, err
	}

	_, ok := data["target-ids"]
	if !ok {
		return nil, errors.New("body missing target-ids field")
	}

	retval := data["target-ids"].([]string)
	return retval, nil
}
