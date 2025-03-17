package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

type HttpClusterAdapter struct{}

func (a HttpClusterAdapter) isError(resp *http.Response) error {
	if resp.StatusCode == 200 {
		return nil
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType == "application/json" {
			defer resp.Body.Close()
			var r contract.ErrorResponse
			err := json.NewDecoder(resp.Body).Decode(&r)
			if err != nil {
				return fmt.Errorf("response format error: %v", err)
			}
			return errors.New(r.Error)
		}
	}

	return fmt.Errorf("httpcode %v", resp.StatusCode)
}

func (a HttpClusterAdapter) Add(
	url string,
	group string,
	kind string,
	param map[string]string,
) (id string, err error) {
	r := contract.AddRequest{
		Group: group,
		Kind:  kind,
		Param: param,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return id, fmt.Errorf("request format error: %v", err)
	}

	resp, err := http.Post(url+"/task", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return id, fmt.Errorf("request url %v error: %v", url, err)
	}

	var res contract.AddResponse
	json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return id, fmt.Errorf("response format error: %v", err)
	}

	id = res.Id

	return id, err
}

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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}

	return err
}

func (a HttpClusterAdapter) OwnerReg(url string, owner string, kinds []string) (err error) {
	r := contract.OwnerRegRequest{
		Owner:    owner,
		Kinds:    kinds,
		Internal: true,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("request format error: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, url+"/owner", bytes.NewBuffer(json_data))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("create request error: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}

	return err
}

func (a HttpClusterAdapter) SetOffset(url string, owner string, kind string, startId string) (err error) {
	r := contract.SetOffsetRequest{
		Owner:    owner,
		Kind:     kind,
		StartId:  startId,
		Internal: true,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("request format error: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, url+"/offset", bytes.NewBuffer(json_data))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("create request error: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}
	defer resp.Body.Close()

	err = a.isError(resp)
	if err != nil {
		return fmt.Errorf("request url %v error: %v", url, err)
	}

	return err
}

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

func (a HttpClusterAdapter) Pool(
	url string,
	owner string,
	kind string,
) (tasks []contract.Task, err error) {
	r := contract.PoolRequest{
		Owner:    owner,
		Kind:     kind,
		Internal: true,
	}

	json_data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("request format error: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url+"/pool", bytes.NewBuffer(json_data))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, fmt.Errorf("create request error: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
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
