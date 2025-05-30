package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) SearchDeleteErrorTask(
	url string,
	condition *contract.Condition,
	kind *string,
	size *uint,
) (err error) {
	r := contract.SearchTaskRequest{
		Condition: condition,
		Kind:      kind,
		Size:      size,
		Internal:  true,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("request format error: %v", err)
	}

	resp, err := http.Post(url+"/error/search/delete", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}

	return nil
}
