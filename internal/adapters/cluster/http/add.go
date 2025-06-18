package http

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) Add(
	url string,
	group string,
	kind string,
	owner *string,
	param map[string]string,
) (id string, err error) {
	r := contract.AddRequest{
		Group: group,
		Kind:  kind,
		Owner: owner,
		Param: param,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return id, fmt.Errorf("request format error: %v", err)
	}

	resp, err := a.client.Post(url+"/task", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}

	var res contract.AddResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return id, fmt.Errorf("response format error: %v", err)
	}

	id = res.Id

	return id, err
}
