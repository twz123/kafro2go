package schemaregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	acceptHeaderName = "Accept"
	acceptMediaType  = "application/vnd.schemaregistry.v1+json"

	schemaEndpint = "/schemas/ids/%d"
)

type restAPI struct {
	host string
}

func (api *restAPI) GetSchemaByID(id int64) (string, error) {
	var schemaString struct {
		Schema string `json:"schema"`
	}

	request := &http.Request{
		Header: make(http.Header),
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   api.host,
			Path:   fmt.Sprintf(schemaEndpint, id),
		},
	}
	request.Header.Set(acceptHeaderName, acceptMediaType)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}

	err = json.NewDecoder(resp.Body).Decode(&schemaString)
	if err != nil {
		return "", err
	}

	return schemaString.Schema, nil
}
