package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (a HttpClusterAdapter) GetFirstInGroup(
	url string,
	group string,
) (id string, err error) {
	resp, err := http.Get(url + "/task/group/" + group)
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}

	var res contract.GetFirstInGroupResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return id, fmt.Errorf("response format error: %v", err)
	}

	id = res.Id

	return id, err
}
