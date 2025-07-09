package http

import (
	"encoding/json"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) GetFirstInGroup(
	url string,
	group string,
) (id string, err error) {
	resp, err := a.client.Get(url + "/task/group/" + group)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}

	err = a.isError(resp)
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}

	var res contract.GetFirstInGroupResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return id, fmt.Errorf("response format error: %v", err)
	}

	id = res.Id

	return id, err
}
