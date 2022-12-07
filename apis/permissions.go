package apis

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/samber/lo"
)

type PermissionsAPI struct {
	permissionsURL *url.URL
}

func NewPermissionsAPI(permissionsURL *url.URL) *PermissionsAPI {
	return &PermissionsAPI{
		permissionsURL: permissionsURL,
	}
}

type Permission struct {
	ResourceName string `json:"resource_name"`
}
type PermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

func (p *PermissionsAPI) GetPublicIDS(publicGroup string) ([]string, error) {
	fullURL := *p.permissionsURL
	fullURL = *fullURL.JoinPath("permissions", "abbreviated", "subjects", "group", publicGroup, "app")
	resp, err := http.Get(fullURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("status code was not 200")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var body PermissionsResponse
	if err = json.Unmarshal(b, &body); err != nil {
		return nil, err
	}
	retval := lo.Map(body.Permissions, func(item Permission, index int) string {
		return item.ResourceName
	})
	return retval, nil
}
