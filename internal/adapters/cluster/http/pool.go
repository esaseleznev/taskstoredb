package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) Pool(
	url string,
	owner string,
	kind string,
) (tasks []contract.Task, err error) {
	resp, err := http.Get(url + "/pool/" + owner + "/kind/" + kind + "?internal=true")
	if err != nil {
		return nil, fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return nil, fmt.Errorf("request url %v error: %v", url, err)
	}

	json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, fmt.Errorf("response format error: %v", err)
	}

	return tasks, err
}
