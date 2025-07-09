package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

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

	req, err := http.NewRequest(http.MethodPatch, url+"/task", bytes.NewBuffer(json_data))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("create request error: %v", err)
	}

	resp, err := a.client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}

	err = a.isError(resp)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}

	return err
}
