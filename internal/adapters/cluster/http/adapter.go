package http

import (
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
