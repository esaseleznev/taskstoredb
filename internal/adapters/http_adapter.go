package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	entity "github.com/esaseleznev/taskstoredb/internal/domain"
)

type HttpClusterAdapter struct{}

func (a HttpClusterAdapter) Add(
	url string,
	group string,
	kind string,
	param map[string]string,
) (id string, err error) {
	type request struct {
		Group string            `json:"g"`
		Kind  string            `json:"k"`
		Param map[string]string `json:"p"`
	}
	type response struct {
		Id string `json:"id"`
	}

	r := request{
		Group: group,
		Kind:  kind,
		Param: param,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return id, fmt.Errorf("request format error: %v", err)
	}

	resp, err := http.Post(url+"/task/add", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}

	defer resp.Body.Close()

	var res response
	json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return id, fmt.Errorf("response format error: %v", err)
	}

	id = res.Id

	return id, err
}

func (a HttpClusterAdapter) Update(
	url string,
	group string,
	id string,
	status entity.Status,
	param map[string]string,
	error *string,
) (err error) {
	type request struct {
		Id     string            `json:"id"`
		Group  string            `json:"g"`
		Status int               `json:"s"`
		Param  map[string]string `json:"p"`
		Error  *string           `json:"e"`
	}

	r := request{
		Id:     id,
		Group:  group,
		Param:  param,
		Status: int(status),
		Error:  error,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("request format error: %v", err)
	}

	resp, err := http.Post(url+"/task/update", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}

	defer resp.Body.Close()

	return err
}
