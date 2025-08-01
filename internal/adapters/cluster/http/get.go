package http

import (
	"encoding/json"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) Get(
	url string,
	group string,
	id string,
) (task *contract.Task, err error) {
	resp, err := a.client.Get(url + "/task/" + id + "/group/" + group)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("request url %v error: %v", url, err)
	}

	err = a.isError(resp)
	if err != nil {
		return nil, fmt.Errorf("request url %v error: %v", url, err)
	}

	err = json.NewDecoder(resp.Body).Decode(task)
	if err != nil {
		return nil, fmt.Errorf("response format error: %v", err)
	}

	return task, err
}
