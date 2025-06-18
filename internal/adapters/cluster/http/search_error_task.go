package http

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) SearchErrorTask(
	url string,
	condition *contract.Condition,
	kind *string,
	size *uint,
) (tasks []contract.Task, err error) {
	r := contract.SearchTaskRequest{
		Condition: condition,
		Kind:      kind,
		Size:      size,
		Internal:  true,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("request format error: %v", err)
	}

	resp, err := a.client.Post(url+"/error/search", "application/json", bytes.NewBuffer(json_data))
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
