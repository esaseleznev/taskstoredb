package http

import (
	"encoding/json"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) Pool(
	url string,
	owner string,
	kind string,
) (tasks []contract.Task, err error) {
	resp, err := a.client.Get(url + "/pool/" + owner + "/kind/" + kind + "?internal=true")
	if err != nil {
		return nil, fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return nil, fmt.Errorf("request url %v error: %v", url, err)
	}

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, fmt.Errorf("response format error: %v", err)
	}

	return tasks, err
}
