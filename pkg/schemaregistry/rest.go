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
	host          string
	cachedSchemas map[int64]string
}

func (api *restAPI) GetSchemaByID(id int64) (string, error) {
	if schema, ok := api.cachedSchemas[id]; ok {
		return schema, nil
	}

	schema, err := api.getSchemaByID(id)
	if err != nil {
		return "", err
	}

	if api.cachedSchemas == nil {
		api.cachedSchemas = make(map[int64]string, 16)
	}

	api.cachedSchemas[id] = schema
	return schema, nil
}

func (api *restAPI) getSchemaByID(id int64) (string, error) {
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
