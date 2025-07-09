package http

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) SearchUpdateTask(
	url string,
	up contract.TaskUpdate,
	condition *contract.Condition,
	kind *string,
	size *uint,
) (err error) {
	r := contract.SearchUpdateTaskRequest{
		Up:        up,
		Condition: condition,
		Kind:      kind,
		Size:      size,
		Internal:  true,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("request format error: %v", err)
	}

	resp, err := a.client.Post(url+"/task/search/update", "application/json", bytes.NewBuffer(json_data))
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

	return nil
}
