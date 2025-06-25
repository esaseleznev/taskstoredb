package leveldb

import (
	"encoding/json"
	"fmt"

	common "github.com/esaseleznev/taskstoredb/internal/adapters/store/common"
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (l LevelAdapter) UpdateError(
	id string,
	status contract.Status,
	param map[string]string,

) (events []contract.Event, err error) {
	taskError, err := l.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get tast from db error: %v", err)
	}
	if taskError == nil {
		return
	}

	taskError.Param = param
	taskError.Status = status

	payload := common.NewPlayload()

	switch status {
	case contract.FAILED:
		taskBytes, err := json.Marshal(taskError)
		if err != nil {
			return nil, fmt.Errorf("task marshal error: %v", err)
		}
		payload.Put([]byte(id), taskBytes)
	case contract.VIRGIN:
	case contract.SCHEDULED:
		task, idNew, keyGroup, err := l.newTask(taskError.Group, taskError.Kind, taskError.Owner, taskError.Param)
		if err != nil {
			return nil, err
		}

		taskBytes, err := json.Marshal(task)
		if err != nil {
			return nil, fmt.Errorf("taskNew marshal error: %v", err)
		}

		payload.Put([]byte(idNew), []byte(taskBytes))
		payload.Put([]byte(keyGroup), []byte(idNew))
		payload.Delete([]byte(id), nil)
	case contract.COMPLETED:
		payload.Delete([]byte(id), nil)
	default:
		return nil, fmt.Errorf("unexpected status: %v", status)
	}

	return payload.Data(), err
}
