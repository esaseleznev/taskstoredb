package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

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

	req, err := http.NewRequest(http.MethodPut, url+"/owner/reg", bytes.NewBuffer(json_data))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("create request error: %v", err)
	}

	resp, err := a.client.Do(req)
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

	return err
}
