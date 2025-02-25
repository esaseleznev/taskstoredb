package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

type HttpClusterAdapter struct{}

func (a HttpClusterAdapter) isError(resp *http.Response) error {
	if resp.StatusCode == 200 {
		return nil
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType == "application/json" {
			defer resp.Body.Close()
			var r contract.ErrorResponse
			err := json.NewDecoder(resp.Body).Decode(&r)
			if err != nil {
				return fmt.Errorf("response format error: %v", err)
			}
			return errors.New(r.Error)
		}
	}

	return fmt.Errorf("httpcode %v", resp.StatusCode)
}

func (a HttpClusterAdapter) Add(
	url string,
	group string,
	kind string,
	param map[string]string,
) (id string, err error) {
	r := contract.AddRequest{
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

	err = a.isError(resp)
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}

	var res contract.AddResponse
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
	status contract.Status,
	param map[string]string,
	error *string,
) (err error) {
	r := contract.UpdateRequest{
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

	err = a.isError(resp)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}

	return err
}
